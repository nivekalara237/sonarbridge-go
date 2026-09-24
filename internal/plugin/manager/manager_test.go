package manager_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"sync"
	"testing"
	"time"
)

type nopRegistry struct{}

func (nopRegistry) Resolve(ctx context.Context, name, v string) (registry.ArtifactRef, error) {
	return registry.ArtifactRef{}, nil
}

func (nopRegistry) Fetch(ctx context.Context, a registry.ArtifactRef, dir string) (string, error) {
	return "", nil
}

var fakeBinaryContent = []byte("fake-binary-content")

func fakeSHA256() string {
	sum := sha256.Sum256(fakeBinaryContent)
	return hex.EncodeToString(sum[:])
}

type fakeRegistry struct{}

func (fakeRegistry) Fetch(ctx context.Context, a registry.ArtifactRef, destDir string) (string, error) {
	path := filepath.Join(destDir, "bin")
	if err := os.WriteFile(path, fakeBinaryContent, 0o755); err != nil {
		return "", err
	}
	return path, nil
}
func (fakeRegistry) Resolve(ctx context.Context, name, versionConstraint string) (registry.ArtifactRef, error) {
	return registry.ArtifactRef{
		Name:            name,
		Version:         "1.0.0",
		ProtocolVersion: "1",
		SHA256:          fakeSHA256(),
	}, nil
}

func newFakeAdapterFactory() func(info runtime.InstanceInfo) runtime.Adapter {
	return func(info runtime.InstanceInfo) runtime.Adapter {
		return runtime.NewFakeAdapter(runtime.InstanceInfo{
			Name: "fake-vcs", Version: "1.0.0", PluginType: "vcs",
			ProtocolVersion: "1", Capabilities: []string{"repository"},
		})
	}
}

func writeFakePlugin(t *testing.T, pluginsDir, name string) {
	t.Helper()
	dir := filepath.Join(pluginsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := map[string]any{
		"name": name, "version": "0.0.1", "type": "vcs",
		"protocolVersion": "1", "capabilities": []string{"repository"},
		"binary": "bin", "sha256": "deadbeef",
	}

	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bin"), []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestManagerStartsUnknownPlugin(t *testing.T) {
	root := t.TempDir()
	st := state.NewStore(filepath.Join(root, "state.json"))
	m := manager.New(filepath.Join(root, "plugins"), st, nopRegistry{}, nil)

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if err := m.Start(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected error starting an undiscovered plugin")
	}
}

func TestManagerFullCycle(t *testing.T) {
	root := t.TempDir()
	pluginsDir := filepath.Join(root, "plugins")
	writeFakePlugin(t, pluginsDir, "fake-vcs")

	st := state.NewStore(filepath.Join(root, "state.json"))
	if err := st.Save(map[string]state.PluginRecord{
		"fake-vcs": {Name: "fake-vcs", Version: "0.0.1", Enabled: true},
	}); err != nil {
		t.Fatal(err)
	}

	m := manager.New(pluginsDir, st, nopRegistry{}, newFakeAdapterFactory())

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if err := m.Start(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("start: %v", err)
	}

	statuses := m.List()
	if len(statuses) != 1 || statuses[0].State != lifecycle.StateReady {
		t.Fatalf("expected fake-vcs READY, got %+v", statuses)
	}
	if statuses[0].Info.ProtocolVersion != "1" || len(statuses[0].Info.Capabilities) != 1 {
		t.Fatalf("handshake info not propagated: %+v", statuses[0].Info)
	}

	if err := m.Stop(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	statuses = m.List()
	if statuses[0].State != lifecycle.StateStopped {
		t.Fatalf("expected STOPPED, got %s", statuses[0].State)
	}
}

func TestManagerInstallEnableDisableUninstall(t *testing.T) {
	root := t.TempDir()
	st := state.NewStore(filepath.Join(root, "state.json"))
	m := manager.New(filepath.Join(root, "plugins"), st, fakeRegistry{}, newFakeAdapterFactory())

	if err := m.Install(context.Background(), "fake-vcs", "1.0.0"); err != nil {
		t.Fatalf("install: %v", err)
	}
	if got := m.List(); len(got) != 1 || got[0].State != lifecycle.StateDisabled {
		t.Fatalf("expected DISABLED right after install, got %+v", got)
	}

	if err := m.Enable("fake-vcs"); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if err := m.Start(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if got := m.List(); got[0].State != lifecycle.StateReady {
		t.Fatalf("expected READY, got %s", got[0].State)
	}

	// Disable while running must stop it first, then flip to DISABLED —
	// and persist that choice (checked via a fresh Manager reading the
	// same state file).
	if err := m.Disable(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if got := m.List(); got[0].State != lifecycle.StateDisabled {
		t.Fatalf("expected DISABLED after disable, got %s", got[0].State)
	}

	records, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if records["fake-vcs"].Enabled {
		t.Fatalf("expected persisted record to be disabled, got %+v", records["fake-vcs"])
	}

	if err := m.Uninstall(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if got := m.List(); len(got) != 0 {
		t.Fatalf("expected no plugins after uninstall, got %+v", got)
	}
	records, err = st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := records["fake-vcs"]; exists {
		t.Fatalf("expected record removed from state after uninstall")
	}
}

func TestManagerCrashTriggersRetryThenReady(t *testing.T) {
	root := t.TempDir()
	pluginsDir := filepath.Join(root, "plugins")
	writeFakePlugin(t, pluginsDir, "fake-vcs")

	st := state.NewStore(filepath.Join(root, "state.json"))
	if err := st.Save(map[string]state.PluginRecord{
		"fake-vcs": {Name: "fake-vcs", Version: "0.0.1", Enabled: true},
	}); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var adapters []*runtime.FakeAdapter

	m := manager.New(pluginsDir, st, nopRegistry{}, func(runtime.InstanceInfo) runtime.Adapter {
		a := runtime.NewFakeAdapter(runtime.InstanceInfo{Name: "fake-vcs", ProtocolVersion: "1"})
		mu.Lock()
		adapters = append(adapters, a)
		mu.Unlock()
		return a
	})
	// Fast backoff for the test instead of the production 1s/2s/4s/...
	m.RestartBackoff = []time.Duration{time.Millisecond, time.Millisecond}

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if err := m.Start(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("start: %v", err)
	}

	mu.Lock()
	first := adapters[0]
	mu.Unlock()
	first.Crash()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s := m.List()
		mu.Lock()
		n := len(adapters)
		mu.Unlock()
		if len(s) == 1 && s[0].State == lifecycle.StateReady && n == 2 {
			return // crashed -> retried on a fresh adapter -> READY again
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected crash -> retry -> READY within the deadline, got %+v", m.List())
}

func TestManagerCrashGivesUpAfterMaxRetries(t *testing.T) {
	root := t.TempDir()
	pluginsDir := filepath.Join(root, "plugins")
	writeFakePlugin(t, pluginsDir, "fake-vcs")

	st := state.NewStore(filepath.Join(root, "state.json"))
	if err := st.Save(map[string]state.PluginRecord{
		"fake-vcs": {Name: "fake-vcs", Version: "0.0.1", Enabled: true},
	}); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var adapters []*runtime.FakeAdapter

	m := manager.New(pluginsDir, st, nopRegistry{}, func(runtime.InstanceInfo) runtime.Adapter {
		a := runtime.NewFakeAdapter(runtime.InstanceInfo{Name: "fake-vcs", ProtocolVersion: "1"})
		mu.Lock()
		adapters = append(adapters, a)
		mu.Unlock()
		return a
	})
	m.RestartBackoff = []time.Duration{time.Millisecond} // only one retry allowed

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if err := m.Start(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Crash every adapter as soon as it appears, forever exhausting retries.
	go func() {
		seen := 0
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			n := len(adapters)
			mu.Unlock()
			if n > seen {
				mu.Lock()
				a := adapters[n-1]
				mu.Unlock()
				a.Crash()
				seen = n
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s := m.List()
		if len(s) == 1 && s[0].State == lifecycle.StateFailed {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected FAILED after exhausting retries, got %+v", m.List())
}
