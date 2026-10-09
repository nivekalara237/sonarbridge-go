package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sonarbridge-go/internal/plugin/discovery"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"strings"
	"sync"
	"time"
)

// ErrUnknownPlugin is wrapped by every Manager method that takes a plugin
// name when no such plugin is known. Match it with errors.Is.
var ErrUnknownPlugin = errors.New("unknown plugin")

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

var defaultProgramArgs = []string{"start"}

// stagingPrefix marks scratch directories inside the plugins directory.
// Discovery skips every dot-prefixed directory, so a staging directory
// left behind by a crash is never mistaken for a plugin.
const stagingPrefix = ".staging-"

// instance bundles everything the Manager tracks for one plugin, running or not
type instance struct {
	fsm           *lifecycle.FSM
	adapter       runtime.Adapter
	info          runtime.InstanceInfo // set after a successful handshake
	binPath       string
	stopRequested bool              // set by stop/Disable/uninstall so a concurrent crash-watcher backs off
	retries       int               // consecutive crash-restart attempts since the last deliberate Start
	env           map[string]string // extra env for the child process
	programArgs   []string          // extra args for the subprocess
}

type Manager struct {
	mu         sync.Mutex
	pluginsDir string
	store      *state.Store
	storeMu    sync.Mutex
	registry   registry.Client

	newAdapter func(info runtime.InstanceInfo) runtime.Adapter

	RestartBackoff []time.Duration

	instances map[string]*instance

	PluginEnv map[string]map[string]string
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
			fsm:         fsm,
			binPath:     d.BinaryPath,
			programArgs: d.Manifest.Args,
			env:         m.PluginEnv[d.Manifest.Name],
		}
	}

	return nil
}

// Start spawns and handshakes the named plugin (triggered eagerly at boot pr lazily on first use)
// a successful Start resets the crash-retry counter and arms crash supervision for this instance
func (m *Manager) Start(ctx context.Context, name string) error {
	return m.startPlugin(ctx, name, false)
}

func (m *Manager) startPlugin(ctx context.Context, name string, isRetry bool) error {
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
	if err := adapter.Start(ctx, inst.binPath, inst.programArgs, inst.env); err != nil {
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

	// TODO: validate info.ProtocolVersion against [minSupported, maxSupported]
	// before accepting READY - this is where protocol compatibility is enforced

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
	if err := m.startPlugin(context.Background(), name, true); err != nil {
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

// Install resolves, download, verifies (checksum) and registers a plugin. without starting, it.
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

	removeStateStaging(m.pluginsDir, name)
	staging, err := os.MkdirTemp(m.pluginsDir, stagingPrefix+name+"-")
	if err != nil {
		return fmt.Errorf("manager: install %q: %w", name, err)
	}

	defer os.RemoveAll(staging)

	stagedBin, err := m.registry.Fetch(ctx, artifact, staging)
	if err != nil {
		return fmt.Errorf("manager: fetch %sq: %w", name, err)
	}

	/*binPath, err := m.registry.Fetch(ctx, artifact, dir)
	if err != nil {
		return fmt.Errorf("manager: fetch %q: %w", name, err)
	}*/

	if err := m.verifyChecksum(stagedBin, artifact.SHA256); err != nil {
		// _ = os.RemoveAll(dir)
		return fmt.Errorf("manager: install %q: %w", name, err)
	}

	dir = filepath.Join(m.pluginsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("manager: install %q: %w", name, err)
	}

	binPath := filepath.Join(dir, filepath.Base(stagedBin))
	if err := os.Rename(stagedBin, binPath); err != nil {
		return fmt.Errorf("manager: install %q: move binary: %w", name, err)
	}

	manifest := discovery.Manifest{
		Name:            artifact.Name,
		Version:         artifact.Version,
		Type:            "vcs",
		ProtocolVersion: artifact.ProtocolVersion,
		Binary:          filepath.Base(binPath),
		SHA256:          artifact.SHA256,
		// Capabilities are intentionally left empty here. It's only authoritative once the
		// plugin declares it at handshake time.
		Args: make([]string, 0), // TODO: populate this field by calling the right way (Important)
	}

	data, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return fmt.Errorf("manager: install %q: %w", name, err)
	}
	if err := writeFileAtomic(filepath.Join(dir, "plugin.json"), data, 0o644); err != nil {
		return fmt.Errorf("manager: install %q: write manifest: %w", name, err)
	}

	fsm := lifecycle.NewFSM()
	_ = fsm.Transition(lifecycle.StateInstalled)
	_ = fsm.Transition(lifecycle.StateDisabled)

	m.mu.Lock()
	m.instances[name] = &instance{fsm: fsm, binPath: binPath, env: m.PluginEnv[name], programArgs: defaultProgramArgs}
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

// Restart stops the plugin if it is running, then starts it again. From STOPPED, FAILED, CRASHED and ENABLED
// there is nothing to stop, and it is a plain Start; from any other state (DISABLED, a start or stop already
// in flight...) the FSM refuses the transition and the error wraps lifecycle.ErrIllegalTransition. If the stop
// succeeds but the new start fails, the plugin is left FAILED — the state tells the truth.

func (m *Manager) Restart(ctx context.Context, name string) error {
	m.mu.Lock()
	inst, ok := m.instances[name]
	running := ok && inst.fsm.Current() == lifecycle.StateReady
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("manager: %w %q", ErrUnknownPlugin, name)
	}

	if running {
		if err := m.Stop(ctx, name); err != nil {
			return err
		}
	}
	return m.Start(ctx, name)
}

// Dispense returns the client-side stub for a business service (see plugin.*Key constants in cli-bridge-plugin-sdk)
// on an already READY plugin. Ex: CommentAndNoteService once Handshake's
// Capabilities confirmed the plugin declares it. The caller is expected to have checked capabilities
// first; Dispense itself doesn't fail just because the plugin didn't register that key, the real RPC
// call on the returned client will, with 	 gRPC "Unimplemented" error.
func (m *Manager) Dispense(name, key, requiredCapability string) (any, error) {
	m.mu.Lock()
	inst, ok := m.instances[name]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("manager: unknown plugin %q", name)
	}

	m.mu.Lock()
	adapter := inst.adapter
	fsmState := inst.fsm.Current()
	caps := inst.info.Capabilities
	m.mu.Unlock()

	if fsmState != lifecycle.StateReady {
		return nil, fmt.Errorf("manager: %q is not READY (state=%s)", name, fsmState)
	}

	if adapter == nil {
		return nil, fmt.Errorf("manager: %q is not started", name)
	}

	if requiredCapability != "" && !hasCapability(caps, requiredCapability) {
		return nil, fmt.Errorf("manager: %q does not declare capability %q(has %v)", name, requiredCapability, caps)
	}

	d, ok := adapter.(runtime.Dispenser)
	if !ok {
		return nil, fmt.Errorf("manager: adapter for %q does not support Dispense", name)
	}

	return d.Dispense(key)
}

func (m *Manager) transition(inst *instance, next lifecycle.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return inst.fsm.Transition(next)
}

func (m *Manager) persistRecord(name, version string, enabled bool) error {
	return m.updateRecords(func(records map[string]state.PluginRecord) {
		records[name] = state.PluginRecord{Name: name, Version: version, Enabled: enabled}
	})
}

func (m *Manager) updateRecords(change func(map[string]state.PluginRecord)) error {
	m.storeMu.Lock()
	defer m.storeMu.Unlock()

	records, err := m.store.Load()
	if err != nil {
		return err
	}

	change(records)
	return m.store.Save(records)
}

func hasCapability(capa []string, want string) bool {
	return slices.Contains(capa, want)
}

func (m *Manager) setEnabled(name string, enabled bool) error {
	return m.updateRecords(func(records map[string]state.PluginRecord) {
		rec, ok := records[name]
		if !ok {
			rec = state.PluginRecord{Name: name}
		}
		rec.Enabled = enabled
		records[name] = rec
	})
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

// writeFileAtomic writes via a temp file in the same directory and renames
// it into place, so a reader never sees a half-written file.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}

func removeStateStaging(pluginDir, name string) {
	stale, _ := filepath.Glob(filepath.Join(pluginDir, stagingPrefix+name+"-*"))
	for _, d := range stale {
		_ = os.RemoveAll(d)
	}
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
