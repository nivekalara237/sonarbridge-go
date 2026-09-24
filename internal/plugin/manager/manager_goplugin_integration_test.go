package manager_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"testing"
	"time"
)

func TestManagerFullCycleWithRealGoPluginAdapter(t *testing.T) {
	repoRoot := findRepoRoot(t)
	binSrc := buildFakePluginBinary(t, repoRoot)

	root := t.TempDir()
	pluginsDir := filepath.Join(root, "plugins")
	pluginDir := filepath.Join(pluginsDir, "fake-vcs")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	binDst := filepath.Join(pluginDir, "fakevcsplugin")
	copyFile(t, binSrc, binDst, 0o755)

	manifest := map[string]any{
		"name": "fake-vcs", "version": "0.0.1", "type": "vcs",
		"protocolVersion": "1", "capabilities": []string{"repository"},
		"binary": "fakevcsplugin", "sha256": "unchecked-in-this-test",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	st := state.NewStore(filepath.Join(root, "state.json"))
	if err := st.Save(map[string]state.PluginRecord{
		"fake-vcs": {Name: "fake-vcs", Version: "0.0.1", Enabled: true},
	}); err != nil {
		t.Fatal(err)
	}

	m := manager.New(pluginsDir, st, nil, func(runtime.InstanceInfo) runtime.Adapter {
		return runtime.NewGoPluginAdapter()
	})

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := m.Start(ctx, "fake-vcs"); err != nil {
		t.Fatalf("start: %v", err)
	}

	statuses := m.List()
	if len(statuses) != 1 || statuses[0].State != lifecycle.StateReady {
		t.Fatalf("expected fake-vcs READY, got %+v", statuses)
	}
	if statuses[0].Info.Name != "fake-vcs" || statuses[0].Info.ProtocolVersion != "1" {
		t.Fatalf("handshake info not propagated: %+v", statuses[0].Info)
	}

	if err := m.Stop(context.Background(), "fake-vcs"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if s := m.List(); s[0].State != lifecycle.StateStopped {
		t.Fatalf("expected STOPPED, got %s", s[0].State)
	}
}

func buildFakePluginBinary(t *testing.T, repoRoot string) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "fakevcsplugin")
	cmd := exec.Command("go", "build", "-o", bin, "sonarbridge-go/cmd/fakevcsplugin")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fakevcsplugin: %v\n%s", err, out)
	}
	return bin
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test working directory")
		}
		dir = parent
	}
}

func copyFile(t *testing.T, src, dst string, mode os.FileMode) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		t.Fatal(err)
	}
}
