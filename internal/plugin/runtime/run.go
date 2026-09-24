package runtime

import "context"

type InstanceInfo struct {
	Name            string
	Version         string
	PluginType      string
	ProtocolVersion string
	Capabilities    []string
}

// Adapter is implemented once per transport. The Manager only ever talks
// to this interface, never to go-plugin or gRPC types directly.
type Adapter interface {
	// Start spawns the plugin process (or fake equivalent) at path and
	// blocks until the transport is reachable (but before any business
	// handshake happens)
	Start(ctx context.Context, path string) error

	Handshake(ctx context.Context) (InstanceInfo, error)

	Alive() bool

	// Exited returns a channel when the plugin process termines, mirroring
	// go-plugin's Client.Existed() pattern. The manager watches this to drive
	// the READY ->  CRASHED transition.
	Exited() <-chan struct{}

	Stop(ctx context.Context) error

	Pid() int
}
