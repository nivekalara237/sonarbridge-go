package pluginshared

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
	"google.golang.org/grpc"
)

var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SBRIDE_PLUGIN",
	MagicCookieValue: "vcs",
}

const PluginKey = "vcs"

type PluginInfoGRPCPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
	Impl pluginv1.PluginInfoServer
}

func (p *PluginInfoGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	pluginv1.RegisterPluginInfoServer(s, p.Impl)
	return nil
}

func (p *PluginInfoGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pluginv1.NewPluginInfoClient(c), nil
}

func Map(impl pluginv1.PluginInfoServer) map[string]goplugin.Plugin {
	return map[string]goplugin.Plugin{
		PluginKey: &PluginInfoGRPCPlugin{
			Impl: impl,
		},
	}
}
