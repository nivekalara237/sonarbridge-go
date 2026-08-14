package loader

import (
	"context"
	"fmt"
	"io"
)

type Chain struct {
	loaders []AbstractLoader
}

func NewChain(loaders ...AbstractLoader) *Chain {
	return &Chain{loaders: loaders}
}

func (c *Chain) Supports(ref Ref) bool {
	for _, l := range c.loaders {
		if l.Supports(ref) {
			return true
		}
	}
	return false
}

func (c *Chain) Load(ctx context.Context, ref Ref) (io.ReadCloser, error) {
	for _, l := range c.loaders {
		if !l.Supports(ref) {
			continue
		}
		return l.Load(ctx, ref)
	}

	return nil, fmt.Errorf("%w: %s", ErrNotSupported, ref)
}

func (c *Chain) Stamp(ctx context.Context, ref Ref) (Stamp, error) {
	for _, l := range c.loaders {
		if !l.Supports(ref) {
			continue
		}
		s, ok := l.(Stamper)
		if !ok {
			return Stamp{}, ErrNotSupported
		}
		return s.Stamp(ctx, ref)
	}
	return Stamp{}, fmt.Errorf("%w: %s", ErrNotSupported, ref)
}
