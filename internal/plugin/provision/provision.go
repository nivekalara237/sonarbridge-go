package provision

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sonarbridge-go/internal/plugin/discovery"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
)

type Desired struct {
	Name    string
	Version string
}

// Manager this a part of manager.Manager that provisions drives
type Manager interface {
	Install(ctx context.Context, name, version string) error
	Enable(name string) error
	Start(ctx context.Context, name string) error
	List() []manager.Status
}

type Stage string

const (
	StageInstall Stage = "install"
	StageEnable  Stage = "enable"
	StageStart   Stage = "start"
)

// Result is the outcome for one desired plugin. A non-nil Err means the plugin is degraded; Stage says at which step.
type Result struct {
	Name      string
	Version   string
	Installed bool
	Stage     Stage
	Err       error
}

type Provisioner struct {
	PluginDir      string
	Manager        Manager
	InstallTimeout time.Duration
	Log            *slog.Logger
}

// Run provisions every desired plugin, in order, and returns one Result each.
// It never stops at a failure: that is the degraded mode.

func (p *Provisioner) Run(ctx context.Context, desired []Desired) []Result {
	results := make([]Result, 0, len(desired))

	for _, d := range desired {
		results = append(results, p.provision(ctx, d))
	}

	return results
}

// Orphans lists plugins present on disk that the configuration no longer
// asks for. They are left alone, removing a binary is not startup's call and
// so the caller can warn about them.
func (p *Provisioner) Orphans(desired []Desired) []string {
	want := make(map[string]struct{}, len(desired))

	for _, d := range desired {
		want[d.Name] = struct{}{}
	}

	found, _ := discovery.Scan(p.PluginDir)
	var orphans []string
	for _, f := range found {
		if _, ok := want[f.Manifest.Name]; !ok {
			orphans = append(orphans, f.Manifest.Name)
		}
	}

	return orphans
}

func (p *Provisioner) provision(ctx context.Context, d Desired) Result {
	res := Result{Name: d.Name, Version: d.Version}
	log := p.log().With("plugin", d.Name, "Version", d.Version)

	if reason := p.installReason(d); reason != "" {
		log.Info("installing plugin", "reason", reason)

		ictx, cancel := context.WithTimeout(ctx, p.installTimeout())
		err := p.Manager.Install(ictx, d.Name, d.Version)

		cancel()

		if err != nil {
			log.Error("plugin install failed", "error", err)
			res.Stage, res.Err = StageInstall, err
			return res
		}
		res.Installed = true
	} else {
		log.Info("plugin already installed, no download")
	}

	state, ok := p.stateOf(d.Name)
	if !ok {
		res.Stage, res.Err = StageEnable, fmt.Errorf("provison: %q is unknown to the manager", d.Name)
		return res
	}

	if state == lifecycle.StateDisabled {
		if err := p.Manager.Enable(d.Name); err != nil {
			log.Error("plugin enable failed", "error", err)
			res.Stage, res.Err = StageEnable, err
			return res
		}
		state = lifecycle.StateEnabled
	}

	if state == lifecycle.StateEnabled {
		if err := p.Manager.Start(ctx, d.Name); err != nil {
			log.Error("plugin start failed", "error", err)
			res.Stage, res.Err = StageStart, err
			return res
		}
	}

	return res
}

func (p *Provisioner) installReason(d Desired) string {
	version, problem := inspectInstalled(p.PluginDir, d.Name)
	if problem != "" {
		return problem
	}

	if !sameVersion(version, d.Version) {
		return fmt.Sprintf("installed version is %s", version)
	}

	return ""
}

func (p *Provisioner) stateOf(name string) (lifecycle.State, bool) {
	for _, s := range p.Manager.List() {
		if s.Name == name {
			return s.State, true
		}
	}

	return "", false
}

func (p *Provisioner) log() *slog.Logger {
	if p.Log != nil {
		return p.Log
	}
	return slog.Default()
}

func (p *Provisioner) installTimeout() time.Duration {
	if p.InstallTimeout > 0 {
		return p.InstallTimeout
	}

	return 2 * time.Second
}

func sameVersion(a, b string) bool {
	return strings.TrimPrefix(strings.TrimSpace(a), "v") == strings.TrimPrefix(strings.TrimSpace(b), "v")
}

// inspectInstalled reads what is on disk for name. It returns the installed
// version, or a human-readable problem when the plugin is absent or cannot
// be trusted as it stands. "Trusted" means the binary still hashes to the
// checksum recorded at install : a truncated download, a crash between the
// binary and manifest moves, or a modified binary all fail that check and
// are treated like an absent plugin, so startup converges instead of
// executing something unverified.
func inspectInstalled(pluginDir, name string) (version, problem string) {
	dir := filepath.Join(pluginDir, name)

	data, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return "", "not installed"
	}

	if err != nil {
		return "", fmt.Sprintf("manifest unreadable: %v", err)
	}

	var mf discovery.Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return "", fmt.Sprintf("manifest invalid: %v", err)
	}

	switch {
	case mf.Name != name:
		return "", fmt.Sprintf("manifest names %q, expected %q", mf.Name, name)
	case mf.Binary == "" || mf.Binary == "." || mf.Binary == ".." || strings.ContainsAny(mf.Binary, `/\`):
		return "", "manifest binary must be a bare file name"
	case mf.SHA256 == "":
		return "", "manifest records no checksum"
	}

	sum, err := fileSHA256(filepath.Join(dir, mf.Binary))
	if errors.Is(err, fs.ErrNotExist) {
		return "", "binary missing"
	}

	if err != nil {
		return "", fmt.Sprintf("binary unreadable: %v", err)
	}

	if !strings.EqualFold(sum, mf.SHA256) {
		return "", "binary does not match its recorded checksum"
	}

	return mf.Version, ""
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
