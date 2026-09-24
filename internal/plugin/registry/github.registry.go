package registry

import (
	"context"
	"sonarbridge-go/pkg/httpclient"
)

type GithubRegistry struct {
	Owner  string
	Prefix string
	Token  string

	httpClient *httpclient.Client
}

func (r *GithubRegistry) Resolve(ctx context.Context, name, versionConstraint string) (ArtifactRef, error) {
	return ArtifactRef{}, nil
}

func (r *GithubRegistry) Fetch(ctx context.Context, artifact ArtifactRef, destDir string) (string, error) {
	return "", nil
}
