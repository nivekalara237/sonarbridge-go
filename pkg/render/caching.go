package render

import (
	"context"
)

type Cache interface {
	Get(ctx context.Context, key string) (Template, bool)
	Set(ctx context.Context, key string, t Template) error
	Invalidate(ctx context.Context, key string) error
}

// noopCache is default when no cache is configured

type noopCache struct{}

func (noopCache) Get(context.Context, string) (Template, bool) {
	return nil, false
}
func (noopCache) Set(context.Context, string, Template) error {
	return nil
}
func (noopCache) Invalidate(context.Context, string) error {
	return nil
}
