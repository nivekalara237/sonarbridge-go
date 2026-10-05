package registry

import "context"

type disableRegistry struct {}

func (disableRegistry) Resolve(_ context.Context, name, _ string) (ArtifactRef, error) {
	return ArtifactRef{}, fmt.Errorf("plugin %q is not installed and plugin.remote_registry is disabled: install it manually or enable yhe registry", name)
}

func (disableRegistry) Fetch(context.Context, ArtifactRef, string) (string, error) {
	return "", errors.New("remote registry is disabled")
}