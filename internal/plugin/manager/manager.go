package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sonarbridge-go/internal/plugin/discovery"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"strings"
	"sync"
	"time"
)

// defaultRestartBackoff is the bounded exponential backoff applied
// between automatic restarts after a crash, per the decision "bounded backoff exponential"
// Overridable via Manager.RestartBackoff (tests use a much shorter schedule).
var defaultRestartBackoff = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	16 * time.Second,
}

// instance bundles everything the Manager tracks for one plugin, running or not

type instance struct {
	fsm           *lifecycle.FSM
	adapter       runtime.Adapter
	info          runtime.InstanceInfo // set after a successful handshake
	binPath       string
	stopRequested bool // set by stop/Disable/uninstall so a concurrent crash-watcher backs off
	retries       int  // consecutive crash-restart attempts since the last deliberate Start
}

type Manager struct {
	mu         sync.Mutex
	pluginsDir string
	store      *state.Store
	registry   registry.Client

	newAdapter func(info runtime.InstanceInfo) runtime.Adapter

	RestartBackoff []time.Duration

	instances map[string]*instance
}

func New(pluginsDir string, store *state.Store, reg registry.Client, newAdapter func(info runtime.InstanceInfo) runtime.Adapter) *Manager {
	return &Manager{
		pluginsDir:     pluginsDir,
		store:          store,
		registry:       reg,
		newAdapter:     newAdapter,
		RestartBackoff: defaultRestartBackoff,
		instances:      make(map[string]*instance),
	}
}

// Bootstrap re-discovers installed plugins and rebuilds in-memory state
// at startup. It does NOT start any plugin - eager/lazy startup policy
// is a separate, not-yet-modeled concern
func (m *Manager) Bootstrap() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	found, _ := discovery.Scan(m.pluginsDir)
	records, err := m.store.Load()
	if err != nil {
		return fmt.Errorf("manager: bootstrap: load state: %w", err)
	}

	for _, d := range found {
		fsm := lifecycle.NewFSM()
		_ = fsm.Transition(lifecycle.StateInstalled)
		if rec, ok := records[d.Manifest.Name]; ok && rec.Enabled {
			_ = fsm.Transition(lifecycle.StateEnabled)
		} else {
			_ = fsm.Transition(lifecycle.StateDisabled)
		}

		m.instances[d.Manifest.Name] = &instance{
			fsm:     fsm,
			binPath: d.BinaryPath,
		}
	}

	return nil
}

// Start spawns and handshakes the named plugin (triggered eagerly at boot pr lazily on first use)
// a successful Start resets the crash-retry counter and arms crash supervision for this instance
func (m *Manager) Start(ctx context.Context, name string) error {
	return m.StartPlugin(ctx, name, false)
}

func (m *Manager) StartPlugin(ctx context.Context, name string, isRetry bool) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok && !isRetry {
		inst.stopRequested = false
		inst.retries = 0
	}

	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("manager: unknown plugin %q", name)
	}

	if err := m.transition(inst, lifecycle.StateStarting); err != nil {
		return err
	}

	adapter := m.newAdapter(runtime.InstanceInfo{Name: name})
	if err := adapter.Start(ctx, inst.binPath); err != nil {
		_ = m.transition(inst, lifecycle.StateFailed)
		return fmt.Errorf("manager: start %q: %w", name, err)
	}

	if err := m.transition(inst, lifecycle.StateHandshaking); err != nil {
		return err
	}

	info, err := adapter.Handshake(ctx)
	if err != nil {
		_ = m.transition(inst, lifecycle.StateFailed)
		return fmt.Errorf("manager: handshake %q: %w", name, err)
	}

	m.mu.Lock()
	inst.adapter = adapter
	inst.info = info
	m.mu.Unlock()

	if err := m.transition(inst, lifecycle.StateReady); err != nil {
		return err
	}
	go m.watchCrash(name, inst, adapter)
	return nil
}

// watchCrash blocks until adapter reports the process has existed, then decides whether that was a deliberates
// Stop()/Disabled()/Uninstall() or ans unexpected crash. On a crash it drives READY -> CRASHED and retries
// with bounded backoff (m.RestartBackoff) before giving up and moving to FAILED. This os the "détecter un plugin
// qui crash [...] redémarrer" requirement
func (m *Manager) watchCrash(name string, inst *instance, adapter runtime.Adapter) {
	<-adapter.Exited()

	m.mu.Lock()
	stopped := inst.stopRequested
	m.mu.Unlock()

	if stopped {
		return
	}

	if err := m.transition(inst, lifecycle.StateCrashed); err != nil {
		return
	}
	m.mu.Lock()
	attempt := inst.retries
	backoff := m.RestartBackoff
	m.mu.Unlock()

	if attempt >= len(backoff) {
		_ = m.transition(inst, lifecycle.StateFailed)
		return
	}

	time.Sleep(backoff[attempt])
	m.mu.Lock()
	inst.retries++
	m.mu.Unlock()
	if err := m.StartPlugin(context.Background(), name, true); err != nil {
		_ = m.transition(inst, lifecycle.StateFailed)
	}
}

// Stop gracefully stops a running plugin instance and disarms crash supervision for it
// (a deliberate stop must never trigger an automatic restart)
func (m *Manager) Stop(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("manager: unknwon plugin %q", name)
	}

	if err := m.transition(inst, lifecycle.StateStopping); err != nil {
		return err
	}
	m.mu.Lock()
	a := inst.adapter
	m.mu.Unlock()
	if a != nil {
		if err := a.Stop(ctx); err != nil {
			return fmt.Errorf("manager: stop %q: %w", name, err)
		}
	}
	return m.transition(inst, lifecycle.StateStopped)
}

// Install resolves, downloard, verifies (checksum) and registers a plugin. without starting it.
// "bridge" never needs a restart for this to take effect.
func (m *Manager) Install(ctx context.Context, name, versionConstraint string) error {
	artifact, err := m.registry.Resolve(ctx, name, versionConstraint)
	if err != nil {
		return fmt.Errorf("manager: resolve %q: %w", name, err)
	}

	dir := filepath.Join(m.pluginsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("manager: fetch %q: %w", name, err)
	}

	binPath, err := m.registry.Fetch(ctx, artifact, dir)
	if err != nil {
		return fmt.Errorf("manager: fetch %q: %w", name, err)
	}

	if err := m.verifyChecksum(binPath, artifact.SHA256); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("manager: install %q: %w", name, err)
	}

	manifest := discovery.Manifest{
		Name:            artifact.Name,
		Version:         artifact.Version,
		Type:            "vcs",
		ProtocolVersion: artifact.ProtocolVersion,
		Binary:          filepath.Base(binPath),
		SHA256:          artifact.SHA256,
	}

	data, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return fmt.Errorf("manager: install %q: %w", name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0o644); err != nil {
		return fmt.Errorf("manager: install %q: %w", name, err)
	}

	fsm := lifecycle.NewFSM()
	_ = fsm.Transition(lifecycle.StateInstalled)
	_ = fsm.Transition(lifecycle.StateDisabled)

	m.mu.Lock()
	m.instances[name] = &instance{fsm: fsm, binPath: binPath}
	m.mu.Unlock()

	return m.persistRecord(name, artifact.Version, false)
}

func (m *Manager) Uninstall(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if !ok {
		// return fmt.Errorf("manager: unknown %q: %w", name, err)
		inst.stopRequested = true
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("manager: unknown plugin %q", name)
	}

	if inst.fsm.Current() == lifecycle.StateReady {
		if err := m.Stop(ctx, name); err != nil {
			return fmt.Errorf("manager: uninstall %q: %w", name, err)
		}
	}

	if err := os.RemoveAll(filepath.Dir(inst.binPath)); err != nil {
		return fmt.Errorf("manager: uninstall %q: %w", name, err)
	}

	m.mu.Lock()
	delete(m.instances, name)
	m.mu.Unlock()

	records, err := m.store.Load()
	if err != nil {
		return err
	}
	delete(records, name)
	return m.store.Save(records)
}

func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("manager: unknown plugin %q", name)
	}

	if err := m.transition(inst, lifecycle.StateEnabled); err != nil {
		return err
	}
	return m.setEnabled(name, true)
}

func (m *Manager) Disable(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	if ok {
		inst.stopRequested = true
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("manager: unknown plugin %q", name)
	}

	if inst.fsm.Current() == lifecycle.StateReady {
		if err := m.Stop(ctx, name); err != nil {
			return err
		}
	}

	if err := m.transition(inst, lifecycle.StateDisabled); err != nil {
		return err
	}

	return m.setEnabled(name, false)
}

func (m *Manager) transition(inst *instance, next lifecycle.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return inst.fsm.Transition(next)
}

func (m *Manager) persistRecord(name, version string, enabled bool) error {
	records, err := m.store.Load()
	if err != nil {
		return err
	}
	records[name] = state.PluginRecord{Name: name, Version: version, Enabled: enabled}
	return m.store.Save(records)
}

func (m *Manager) setEnabled(name string, enabled bool) error {
	records, err := m.store.Load()
	if err != nil {
		return err
	}
	rec, ok := records[name]
	if !ok {
		rec = state.PluginRecord{Name: name}
	}
	rec.Enabled = enabled
	records[name] = rec
	return m.store.Save(records)
}

func (m *Manager) verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", got, expected)
	}
	return nil
}

type Status struct {
	Name  string
	State lifecycle.State
	Info  runtime.InstanceInfo
}

func (m *Manager) List() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Status, 0, len(m.instances))
	for name, inst := range m.instances {
		out = append(out, Status{Name: name, State: inst.fsm.Current(), Info: inst.info})
	}
	return out
}
