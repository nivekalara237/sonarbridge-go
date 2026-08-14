package render

import (
	"fmt"
	"slices"
	"sync"
)

// Factory build a Renderer. Receive the Engine's options
type Factory func() (Renderer, error)

// type Factory func(opts render.Options) (renderers.Renderer, error)

type Registry interface {
	// Register detect the duplicated on boot time otherwise override silenciously
	Register(format Format, f Factory) error
	// Resolve(format render.Format, opts render.Options) (renderers.Renderer, error)
	Resolve(format Format) (Renderer, error)
	// Formats for introspection, validation of config in the startup
	// Detect potential errors
	Formats() []Format
}

type registry struct {
	mu        sync.RWMutex
	renderers map[Format]Renderer
}

func NewRegistry() Registry {
	return &registry{renderers: make(map[Format]Renderer)}
}

func (r *registry) Register(f Format, factory Factory) error {
	if f.IsZero() {
		return fmt.Errorf("%w: empty form", ErrFormatInvalidRenderer)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.renderers[f]; exists {
		return fmt.Errorf("%w: %s", ErrFormatAlreadyRegistered, f)
	}
	rd, err := factory()
	if err != nil {
		return wrapErr("register", f, "", err)
	}

	if err := validateRoles(f, rd); err != nil {
		return err
	}

	r.renderers[f] = rd
	return nil
}

func validateRoles(f Format, rd Renderer) error {
	_, isCompiler := rd.(Compiler)
	_, isEncoder := rd.(Encoder)
	if !isCompiler && !isEncoder {
		return fmt.Errorf("%w: %s", ErrFormatInvalidRenderer, f)
	}

	if rd.Format() != f {
		return fmt.Errorf("%w: registered as %s but reports %s", ErrFormatInvalidRenderer, f, rd.Format())
	}
	return nil
}

func (r *registry) Resolve(format Format) (Renderer, error) {
	r.mu.RLock()
	rd, ok := r.renderers[format]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFormatNotRegistered, format)
	}
	return rd, nil
}

func (r *registry) Formats() []Format {
	r.mu.RLock()
	out := make([]Format, 0, len(r.renderers))
	for f := range r.renderers {
		out = append(out, f)
	}
	r.mu.RUnlock()

	slices.Sort(out)
	return out
}
