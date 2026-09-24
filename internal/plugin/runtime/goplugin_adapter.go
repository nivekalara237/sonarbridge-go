package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"sonarbridge-go/internal/plugin/pluginshared"
	"sonarbridge-go/internal/plugin/proto/pluginv1"
	"time"

	goplugin "github.com/hashicorp/go-plugin"
)

type GoPluginAdapter struct {
	client     *goplugin.Client
	infoClient pluginv1.PluginInfoClient
	exited     chan struct{}
	cmd        *exec.Cmd // kept so Pid() can report the real subprocess PID
}

func NewGoPluginAdapter() *GoPluginAdapter {
	return &GoPluginAdapter{exited: make(chan struct{})}
}

func (a *GoPluginAdapter) Start(ctx context.Context, path string) error {
	a.cmd = exec.Command(path)
	a.client = goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  pluginshared.Handshake,
		Plugins:          pluginshared.Map(nil),
		Cmd:              a.cmd,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		AutoMTLS:         true,
	})

	rpcClient, err := a.client.Client()
	if err != nil {
		a.client.Kill()
		return fmt.Errorf("goplugin: connect: %w", err)
	}

	raw, err := rpcClient.Dispense(pluginshared.PluginKey)
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
