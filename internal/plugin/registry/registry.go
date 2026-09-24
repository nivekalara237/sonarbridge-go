package registry

import "context"

type ArtifactRef struct {
	Name            string
	Version         string
	ProtocolVersion string
	OS              string
	Arch            string
	DownloadURL     string
	SHA256          string
	SignatureURL    string
}

type Client interface {
	Resolve(ctx context.Context, name, versionConstraint string) (ArtifactRef, error)

	// Fetch download the artifact to destDir and returns the local
	// path to the binary.
	Fetch(ctx context.Context, artifact ArtifactRef, destDir string) (string, error)
}
