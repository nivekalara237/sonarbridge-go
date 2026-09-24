package main

import (
	"context"
	"sonarbridge-go/internal/plugin/pluginshared"

	goplugin "github.com/hashicorp/go-plugin"
	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
)

type fakeVCSServer struct {
	pluginv1.UnimplementedPluginInfoServer
}

func (s *fakeVCSServer) GetInfo(ctx context.Context, req *pluginv1.GetInfoRequest) (*pluginv1.InfoResponse, error) {
	return &pluginv1.InfoResponse{
		Name:            "fake-vcs",
		Version:         "0.0.1",
		PluginType:      "vcs",
		ProtocolVersion: "1",
		Capabilities:    []string{"repository", "upload34"},
	}, nil
}

func main() {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.Map(&fakeVCSServer{}),
		GRPCServer:      goplugin.DefaultGRPCServer,
	})
}
