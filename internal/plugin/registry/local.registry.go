package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/pkg/httpclient"
	"strings"
	"time"
)

type LocalRegistry struct {
	client     *httpclient.Client
	clientOpts []httpclient.Option
}

type localAsset struct {
}

type LocalRegistryOption func(*LocalRegistry)

func WithBasicAuth(username, password string) LocalRegistryOption {
	return func(registry *LocalRegistry) {
		encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		registry.clientOpts = append(registry.clientOpts, httpclient.WithDefaultHeader("Authorization", "Basic "+encoded))
	}
}

func NewLocalRegistry(baseUrl string, opts ...LocalRegistryOption) *LocalRegistry {
	clientOpts := []httpclient.Option{
		httpclient.WithBaseURL(baseUrl),
		httpclient.WithCache(httpclient.NewInMemoryCache()),
		httpclient.WithTimeout(5 * time.Second),
	}
	reg := &LocalRegistry{}
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(reg)
		}
	}
	clientOpts = append(clientOpts, reg.clientOpts...)
	reg.client = httpclient.NewClientHttp(clientOpts...)
	// debug.PrintLn(baseUrl, len(opts), clientOpts, len(clientOpts))
	return reg
}

func (r *LocalRegistry) httpClient() *httpclient.Client {
	if r.client != nil {
		return r.client
	}
	return httpclient.NewClientHttp()
}

func (r *LocalRegistry) Resolve(ctx context.Context, name, versionConstraint string) (ArtifactRef, error) {
	catalog, err := r.fetchCatalogRelease(ctx, name, versionConstraint)
	if err != nil {
		return ArtifactRef{}, err
	}
	goos, goarch := runtime.GOOS, runtime.GOARCH
	assetName := name
	if goos == "windows" {
		assetName += ".exe"
	}

	asset := findAsset(catalog.Plugins, assetName, goos, goarch, false)
	if asset == nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: no asset named %q in release %s", name, versionConstraint, assetName, catalog.Metadata.Name)
	}

	checksumAsset := findAsset(catalog.Plugins, assetName+"_checksums.txt", goos, goarch, true)
	if checksumAsset == nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: release %s has no checksums.txt asset", name, versionConstraint, catalog.Metadata.Name)
	}
	// vcsType := getVcsTypeByName(name)
	checksums, err := r.downloadBytes(ctx, fmt.Sprintf("assets/binary/%s", utils.EncodeURIComponent(checksumAsset.Path)))
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: fetch checksums: %w", name, versionConstraint, err)
	}

	archName := goarch
	if goarch == "amd64" {
		archName = "x86_64"
	}
	sum, err := extractChecksums(checksums, fmt.Sprintf("%s_%s_%s", asset.Name, goos, archName))
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: %w", name, versionConstraint, err)
	}

	metaBytes, err := r.downloadBytes(ctx, fmt.Sprintf("assets/binary/%s", utils.EncodeURIComponent(catalog.Metadata.Path)))
	if err != nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: fetch metadata: %w", name, versionConstraint, err)
	}

	var metadata Metadata
	if err := json.Unmarshal(metaBytes, &metadata); err != nil {
		return ArtifactRef{}, fmt.Errorf("local-registry: %s@%s: decoding metadata: %w", name, versionConstraint, err)
	}

	escapedUrl := fmt.Sprintf("assets/binary/%s", utils.EncodeURIComponent(asset.Path))

	return ArtifactRef{
		Name:        name,
		Version:     metadata.Version,
		OS:          goos,
		Arch:        goarch,
		DownloadURL: escapedUrl,
		SHA256:      sum,
	}, nil
}

// Fetch downloads artifact.DownloadUR (the GitHub/Registry/LocalSever API asset endpoint, set by Resolve)
// to destDir.
func (r *LocalRegistry) Fetch(ctx context.Context, artifact ArtifactRef, destDir string) (string, error) {
	resp, err := r.client.Get(ctx, artifact.DownloadURL, &httpclient.RequestOptions{})
	if err != nil {
		return "", fmt.Errorf("local-registry: %s: %w", artifact.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("local-registry: fetch %s: %d: %s", artifact.Name, resp.StatusCode, body)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("local-registry: fetch %s: %w", artifact.Name, err)
	}

	destPath := filepath.Join(destDir, artifact.Name)
	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("local-registry: fetch %s: %w", artifact.Name, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("local-registry: fetch %s: write: %w", artifact.Name, err)
	}

	if err := f.Chmod(0o755); err != nil {
		return "", fmt.Errorf("local-registry: fetch %s: chmod: %w", artifact.Name, err)
	}

	return destPath, nil

}

func extractChecksums(checksumsFiles []byte, filename string) (string, error) {
	for line := range strings.SplitSeq(string(checksumsFiles), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.ToLower(fields[1]) == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no checksums entry for %q in checksums.txt", filename)
}

func findAsset(entries []Entry, name, os, arch string, isChecksum bool) *Entry {
	for _, e := range entries {
		if isChecksum && e.Type == "Checksum" {
			return &e
		}
		if (e.Name == name || (e.Extra != nil && e.Extra.Binary == name)) &&
			e.Goos == os && e.Goarch == arch {
			return &e
		}
	}
	return nil
}

func (r *LocalRegistry) fetchCatalogRelease(ctx context.Context, name, versionConstraint string) (*Catalog, error) {
	vcsType := getVcsTypeByName(name)
	var url string

	if versionConstraint == "" || versionConstraint == "latest" {
		url = fmt.Sprintf("/%s/artifacts/latest", vcsType)
	} else {
		tag := versionConstraint
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		url = fmt.Sprintf("/%s/artifacts/%s", vcsType, tag)
	}

	resp, err := r.client.Get(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("local-registry: %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local-registry: no release found for %s: %s", name, body)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local-registry: error (%d) %s: %w", resp.StatusCode, name, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("local-registry: %s: read response: %w", name, err)
	}

	var catalogEntries []Entry
	if err := json.Unmarshal(body, &catalogEntries); err != nil {
		return nil, fmt.Errorf("local-registry: %s: decode catalog: %w", name, err)
	}

	var catalog Catalog

	for _, e := range catalogEntries {
		if e.Type == "Metadata" {
			catalog.Metadata = e
		} else {
			catalog.Plugins = append(catalog.Plugins, e)
		}
	}

	return &catalog, nil
}

func getVcsTypeByName(name string) string {
	var vcsType string
	switch {
	case strings.Contains(strings.ToLower(name), "gitlab"):
		vcsType = "gitlab"
	case strings.Contains(strings.ToLower(name), "github"):
		vcsType = "github"
	case strings.Contains(strings.ToLower(name), "bb"):
		vcsType = "bitbucket"
	default:
		vcsType = "gitlab"
	}
	return vcsType
}

func (r *LocalRegistry) downloadBytes(ctx context.Context, url string) ([]byte, error) {
	resp, err := r.client.Get(ctx, url, &httpclient.RequestOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local-registry: %d: %s", resp.StatusCode, body)
	}
	return body, nil
}
