package core

import (
	"context"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/capability"
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
		Capabilities: []string{
			capability.PullRequestCreateComment,
			capability.PullRequestUpdateComment,
			capability.PullRequestDeleteComment,
			capability.PullRequest,
			capability.Artifact,
			capability.Issue,
			capability.Branch,
			capability.Pipeline,
			capability.Repository,
			capability.Webhook,
		},
	}, nil
}
