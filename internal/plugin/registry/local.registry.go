package registry

import (
	"context"
	"fmt"
	"runtime"
	"sonarbridge-go/pkg/httpclient"
	"time"
)

type LocalRegistry struct {
	Client *httpclient.Client
}

type localAsset struct {
}

func NewLocalRegistry(baseUrl string) *LocalRegistry {
	return &LocalRegistry{Client: httpclient.NewClientHttp(
		httpclient.WithBaseURL(baseUrl),
		httpclient.WithCache(httpclient.NewInMemoryCache()),
		httpclient.WithTimeout(5*time.Second),
	)}
}

func (r *LocalRegistry) httpClient() *httpclient.Client {
	if r.Client != nil {
		return r.Client
	}
	return httpclient.NewClientHttp()
}

func (r *LocalRegistry) Resolve(ctx context.Context, name, versionConstraint string) (ArtifactRef, error) {
	release, err := r.fetchCatalogRelease(ctx, name, versionConstraint)
	if err != nil {
		return ArtifactRef{}, err
	}
	goos, goarch := runtime.GOOS, runtime.GOARCH
	assetName := fmt.Sprintf("%s_%s_%s", name, goos, goarch)
}

func (r *LocalRegistry) fetchCatalogRelease(ctx context.Context, name, versionConstraint string) (*Catalog, error) {

}
