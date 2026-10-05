package grpc

import (
	"fmt"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/plugin/manager"
	"strings"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/comment"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/common"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/pullrequest"
)

type VcsClient struct {
	m *manager.Manager
}

func (c *VcsClient) commentAndNoteClient(provider, requireCapability string) (comment.CommentAndNoteServiceClient, error) {

	var providerBinaryName string

	for _, p := range configs.AppConfig.VcsProviders {
		if strings.ToLower(p.Name) == strings.ToLower(provider) {
			providerBinaryName = p.BinaryName
		}
	}

	if providerBinaryName == "" {
		return nil, fmt.Errorf("vcs: %s: this vcs plugin is not present in list of vcsProviders[_].Name", provider)
	}

	raw, err := c.m.Dispense(providerBinaryName, plugin.CommentAndNoteKey, requireCapability)
	if err != nil {
		return nil, fmt.Errorf("vcs: %s: %w", provider, err)
	}

	client, ok := raw.(comment.CommentAndNoteServiceClient)
	if !ok {
		return nil, fmt.Errorf("vcs: %s: dispensed and unexpected type for comment service: %T", provider, raw)
	}

	return client, nil
}

func (c *VcsClient) pullrequestClient(provider, requireCapability string) (pullrequest.PullRequestServiceClient, error) {
	raw, err := c.m.Dispense(provider, plugin.PullrequestKey, requireCapability)
	if err != nil {
		return nil, fmt.Errorf("vcs: %s: %w", provider, err)
	}

	client, ok := raw.(pullrequest.PullRequestServiceClient)
	if !ok {
		return nil, fmt.Errorf("vcs: %s: dispensed and unexpected type for pullrequest service: %T", provider, raw)
	}

	return client, nil
}

func isSuccessStatus(status int32) bool {
	return status >= 200 && status < 300
}

func errMsg(meta *common.HttpResponseMetadata) string {
	if meta.ErrorMsg != nil {
		return *meta.ErrorMsg
	}
	return ""
}
