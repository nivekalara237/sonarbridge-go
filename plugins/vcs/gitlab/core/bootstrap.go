package core

import (
	"context"

	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
)

type GitlabCVSServer struct {
	pluginv1.UnimplementedPluginInfoServer
}

func (s *GitlabCVSServer) GetInfo(ctx context.Context, req *pluginv1.GetInfoRequest) (*pluginv1.InfoResponse, error) {
	return &pluginv1.InfoResponse{
		Name:            "gitlab",
		Version:         "",
		PluginType:      "vcs",
		ProtocolVersion: "1",
		Capabilities:    nil,
	}, nil
}
