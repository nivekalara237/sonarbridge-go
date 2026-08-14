package render

import (
	"bytes"
	"sonarbridge-go/pkg/render/loader"
	"sync"
)

type Options struct {
}

// Option configure the engine. This pattern is preferred than configuration
// The options compose, hold their values by default and adding one more will
// never a breaking change
type Option func(*engine)

// WithLoader replace the loader. Pass a leader.Chain to combine more source
func WithLoader(l loader.AbstractLoader) Option {
	return func(e *engine) {
		e.loader = l
	}
}

func WithRegistry(r Registry) Option {
	return func(e *engine) {
		e.registry = r
	}
}

// WithCache Connect a template Cache compiled. Without it each recompile
func WithCache(c Cache) Option {
	return func(e *engine) {
		e.cache = c
	}
}

func WithDefaultFormat(f Format) Option {
	return func(e *engine) {
		e.defaultFormat = f
	}
}

// WithAtomicWrites make render all-in-one: The output transite by a mutual
// buffer and didn't copy onto w only is success case.
func WithAtomicWrites(enabled bool) Option {
	return func(e *engine) {
		e.atomic = enabled
	}
}

// WithBufferPool inject a shared pool. By default, the Engine create it own.
// The sharing between many Engine(s) reduce the GC pressure
func WithBufferPool(p *sync.Pool) Option {
	return func(e *engine) {
		e.bufPool = p
	}
}

func defaultBufferPool() *sync.Pool {
	return &sync.Pool{New: func() any {
		return new(bytes.Buffer)
	}}
}
