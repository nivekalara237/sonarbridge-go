package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "\nFAILED", err)
		os.Exit(1)
	}
}

func getOeCreateAppDirOLD() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(home, ".sbridge")

	fmt.Println("HOME:", home)
	fmt.Println("DIR :", dir)

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		panic(err)
	}

	// Vérification immédiate
	info, err := os.Stat(dir)
	if err != nil {
		panic(err)
	}

	fmt.Println("Created successfully!")
	fmt.Println("Exists:", true)
	fmt.Println("IsDir:", info.IsDir())

	// Liste le contenu du HOME
	entries, err := os.ReadDir(home)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.Name() == ".sbridge" {
			fmt.Println("FOUND .sbridge in HOME")
		}
	}

	return dir
}
func getOeCreateAppDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(home, ".sbridge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	fmt.Println("App dir initiated: ", dir)

	return dir
}

func run() error {
	// root, err := os.MkdirTemp("", "sonarbridge-e2e-*")
	root := getOeCreateAppDir()
	// fmt.Println(os.Dir("$HOME/.sbridge"))
	// fmt.Println(root)

	// defer os.RemoveAll(root)

	step("1/7", "build fakevcsplugin")
	bin, err := buildFakePlugin(root)
	if err != nil {
		return err
	}

	fmt.Printf("  binary: %s\n", bin)
	step("2/7", "Install (manifest + binary déposés sous plugins/fake-vcs/)")
	pluginsDir := filepath.Join(root, "plugins")
	if err := installFakePlugin(pluginsDir, bin); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	st := state.NewStore(filepath.Join(root, "state.json"))
	if err := st.Save(map[string]state.PluginRecord{
		"fake-vcs": {Name: "fake-cvs", Version: "0.0.1", Enabled: true},
	}); err != nil {
		return err
	}

	var lastAdapter runtime.Adapter
	m := manager.New(pluginsDir, st, nil, func(runtime.InstanceInfo) runtime.Adapter {
		a := runtime.NewGoPluginAdapter()
		lastAdapter = a
		return a
	})

	m.RestartBackoff = []time.Duration{500 * time.Millisecond}

	step("3/7", "discover (Bootstraping)")
	if err := m.Bootstrap(); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	printStatus(m)

	step("4/7", "launch + handshake(vrai subprocess, vraie RPC gRPC)")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := m.Start(ctx, "fake-vcs"); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	printStatus(m)

	pid := lastAdapter.Pid()
	if pid == 0 {
		return fmt.Errorf("pid inconnu juste après Start - inattendu")
	}

	step("5/7", fmt.Sprintf("tuer le vrai process(pid=%s, SIGKILL) pour simuler un crash", pid))

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("kill: %w", err)
	}

	step("6/7", "observer le Manager détecter le crash et redémarrer automatiquement")

	if err := waitForState(m, lifecycle.StateCrashed, 5*time.Second); err != nil {
		return fmt.Errorf("jamais vu CRASHED: %w", err)
	}

	printStatus(m)
	if err := waitForState(m, lifecycle.StateReady, 3*time.Second); err != nil {
		return fmt.Errorf("jamais revenu à READY après le crash: %w", err)
	}

	printStatus(m)
	if newPid := lastAdapter.Pid(); newPid != pid {
		fmt.Printf("    nouveau process après retry : pid=%d (ancien pid=%d, bien mort)\n", newPid, pid)
	}

	step("7/7", "stop")
	if err := m.Stop(context.Background(), "fake-vcs"); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	printStatus(m)

	fmt.Println("\nOK — cycle complet (discover -> launch -> handshake -> crash réel -> retry -> stop) validé avec un vrai subprocess.")

	return nil
}

func step(n, desc string) {
	fmt.Printf("\n[%s] %s\n", n, desc)
}

func printStatus(m *manager.Manager) {
	for _, s := range m.List() {
		fmt.Printf(" status: %s -> %s", s.Name, s.State)
		if s.Info.Name != "" {
			fmt.Printf(" (version=%s protocol=%d capabilities=%v)", s.Info.Version, s.Info.ProtocolVersion, s.Info.Capabilities)
		}
		fmt.Println()
	}

}

func waitForState(m *manager.Manager, want lifecycle.State, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, s := range m.List() {
			if s.Name == "fake-vcs" && s.State == want {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("timeout en attendant l'etat %s", want)
}

func installFakePlugin(pluginSDir, srcBin string) error {
	dir := filepath.Join(pluginSDir, "fake-vcs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(srcBin)
	if err != nil {
		return err
	}

	dst := filepath.Join(dir, "fakevcsplugin")
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		return err
	}

	manifest := map[string]any{
		"name":            "fake-vcs",
		"version":         "0.0.1",
		"type":            "vcs",
		"protocolVersion": "1",
		"binary":          "fakevcsplugin",
		"sha256":          "unchecked-in-this-demo",
		"capabilities":    []string{"repository", "upload"},
	}

	b, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "plugin.json"), b, 0o644)
}

func buildFakePlugin(outDir string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	bin := filepath.Join(outDir, "fakevcsplugin")
	cmd := exec.Command("go", "build", "-o", bin, "sonarbridge-go/cmd/fakevcsplugin")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return bin, nil
}

func debug(m string) {
	fmt.Println("***********************************")
	fmt.Println(m)
	fmt.Println("***********************************")
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod introuvable en remontant depuis %s", dir)
		}

		dir = parent
	}
}
