package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
)

// GoPluginAdapter is the reak Adapter implementation: it spawns the plugin binary
// as a subprocess and speaks gRPC to it over go-plugin's managed connection.
type GoPluginAdapter struct {
	client     *goplugin.Client
	infoClient pluginv1.PluginInfoClient
	rpcClient  goplugin.ClientProtocol
	exited     chan struct{}
	cmd        *exec.Cmd // kept so Pid() can report the real subprocess PID

	Stdout io.Writer
	Stderr io.Writer
}

// defaultArgs is the fallback when Start is called with no args: every
// ci-bridge plugin binary is a small CLI, not a bare go-plugin server
// — invoking it with no arguments prints usage and exits (so a human
// running it by hand gets help instead of an opaque hung process
// waiting on stdin for a handshake it will never receive). "start" is
// the subcommand that actually calls plugin.Serve. This default exists
// so a caller that forgets to pass args (or passes nil deliberately,
// meaning "use the normal convention") doesn't silently reproduce that
// failure mode.
var defaultArgs = []string{"start"}

func NewGoPluginAdapter() *GoPluginAdapter {
	return &GoPluginAdapter{exited: make(chan struct{})}
}

func (a *GoPluginAdapter) Start(ctx context.Context, path string, arguments []string, env map[string]string) error {

	if len(arguments) == 0 {
		arguments = defaultArgs
	}
	a.cmd = exec.Command(path, arguments...)
	if len(env) > 0 {
		// Extend the host's own environment rather than replacement it
		// The plugin still needs PATH, HOME, etc. Any key here overrides the inherited value if
		// it collides (last write wins in os/exec's Env, and these are appended last)
		extra := make([]string, 0, len(env))
		for k, v := range env {
			extra = append(extra, k+"="+v)
		}
		a.cmd.Env = append(os.Environ(), extra...)
	}

	stdout, stderr := a.Stdout, a.Stderr
	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	a.client = goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  plugin.Handshake,
		Plugins:          plugin.HostPlugins(),
		Cmd:              a.cmd,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		// AutoMTLS:         true,
		SkipHostEnv: false,
		SyncStderr:  stderr,
		SyncStdout:  stdout,
		Logger: hclog.New(&hclog.LoggerOptions{
			JSONFormat: false,
			Level:      hclog.Trace,
			Output:     os.Stdout,
		}),
	})

	rpcClient, err := a.client.Client()
	if err != nil {
		fmt.Println(err)
		fmt.Println(a.client.Protocol())
		a.client.Kill()
		return fmt.Errorf("goplugin: connect: %w", err)
	}

	a.rpcClient = rpcClient

	raw, err := rpcClient.Dispense(plugin.PluginKey)
	if err != nil {
		a.client.Kill()
		return fmt.Errorf("goplugin: dispense: %w", err)
	}

	infoClient, ok := raw.(pluginv1.PluginInfoClient)
	if !ok {
		a.client.Kill()
		return fmt.Errorf("goplugin: dispensed value has unexpected type %T", raw)
	}

	a.infoClient = infoClient

	go a.watchExit()
	return nil
}

// watchExit polls client.Exited() (go-plugin expose process exit as a method
// , not a channel) and closes the Adapter's own Exited() channel
// the moment it flips. This is what lets the Manager's watchCrash goroutine react to
// a real crash the same way it reacts to Adapter.Crash(). The 250ms interval is a deliberete
// latency/simplicity tradeoff fpr this increment; watching the underlying os/exex.cmd.Process
// directly would be tighter but couples this adapter to go-plugin.
func (a *GoPluginAdapter) watchExit() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		if a.client.Exited() {
			close(a.exited)
			return
		}
	}
}

func (a *GoPluginAdapter) Handshake(ctx context.Context) (InstanceInfo, error) {
	resp, err := a.infoClient.GetInfo(ctx, &pluginv1.GetInfoRequest{})
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("goplugin: GetInfo: %w", err)
	}

	return InstanceInfo{
		Name:            resp.GetName(),
		Version:         resp.GetVersion(),
		ProtocolVersion: resp.ProtocolVersion,
		Capabilities:    resp.Capabilities,
		PluginType:      resp.PluginType,
	}, nil
}

func (a *GoPluginAdapter) Alive() bool {
	return a.client != nil && !a.client.Exited()
}

func (a *GoPluginAdapter) Exited() <-chan struct{} {
	return a.exited
}

func (a *GoPluginAdapter) Stop(ctx context.Context) error {
	if a.client != nil {
		a.client.Kill()
	}
	return nil
}

func (a *GoPluginAdapter) Pid() int {
	if a.cmd == nil || a.cmd.Process == nil {
		return 0
	}
	return a.cmd.Process.Pid
}

// Dispense return the client-side stub go-plugin registered under key - for calling business
// RPCs beyond the handshake. Only valid after Start has succeeded. The caller is expected to
// have already checked, via Handshake's returned Capabilities, that the plugin actually implements
// what this key maps to.
func (a *GoPluginAdapter) Dispense(key string) (any, error) {
	if a.rpcClient == nil {
		return nil, fmt.Errorf("goplugin: not started")
	}
	return a.rpcClient.Dispense(key)
}
