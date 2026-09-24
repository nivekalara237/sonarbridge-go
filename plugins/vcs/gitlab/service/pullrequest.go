package service

import (
	"context"
	"sonarbridge-go/pkg/httpclient"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/pullrequest"
)

type PullrequestService struct {
	pullrequest.UnimplementedPullRequestServiceServer
	httpclient *httpclient.Client
}

func NewPullrequestService(client *httpclient.Client) *PullrequestService {
	return &PullrequestService{httpclient: client}
}

func (p *PullrequestService) GetPullRequest(ctx context.Context, req *pullrequest.GetPullRequestRequest) (*pullrequest.GetPullRequestResponse, error) {
	return nil, nil
}
