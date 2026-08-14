package render

import (
	"bytes"
	"context"
	"io"
	"sonarbridge-go/pkg/render/loader"
	"sync"
)

type Engine interface {
	// Render : nominal path. Resolve, compile (or cache), Execute
	Render(ctx context.Context, req Request, w io.Writer) error
	RenderString(ctx context.Context, req Request) (string, error)

	// Prewarm : compile at advance (in the boot time, to avoid the cold start).
	Prewarm(ctx context.Context, format Format, refs ...loader.Ref) error

	ContentType(f Format) (string, error)

	Formates() []Format
}

type engine struct {
	loader        loader.AbstractLoader
	registry      Registry
	cache         Cache
	defaultFormat Format
	atomic        bool
	bufPool       *sync.Pool
}

func New(opts ...Option) (Engine, error) {
	e := &engine{
		loader:        loader.NewChain(loader.NewString()),
		registry:      NewRegistry(),
		cache:         noopCache{},
		defaultFormat: FormatMarkdown,
		atomic:        false,
		bufPool:       defaultBufferPool(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

func (e *engine) Render(ctx context.Context, req Request, w io.Writer) error {
	format := req.Format
	if format.IsZero() {
		format = e.defaultFormat
	}

	rd, err := e.registry.Resolve(format)
	if err != nil {
		return wrapErr("resolve", format, req.Ref.String(), err)
	}

	if !e.atomic {
		return e.dispatch(ctx, rd, format, req, w)
	}

	// Atomic mode: Rend in a buffer
	buf := e.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer e.bufPool.Put(buf)

	if err := e.dispatch(ctx, rd, format, req, buf); err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

func (e *engine) RenderString(ctx context.Context, req Request) (string, error) {
	var buff = &bytes.Buffer{}
	err := e.Render(ctx, req, buff)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

func (e *engine) dispatch(ctx context.Context, rd Renderer, format Format, req Request, w io.Writer) error {
	if c, ok := rd.(Compiler); ok {
		tmpl, err := e.template(ctx, c, format, req.Ref)
		if err != nil {
			return err
		}

		return wrapErr("execute", format, req.Ref.String(), tmpl.Execute(ctx, w, req.Data))
	}

	if enc, ok := rd.(Encoder); ok {
		return wrapErr("encode", format, req.Ref.String(), enc.Encode(ctx, w, req.Data))
	}

	return wrapErr("dispach", format, req.Ref.String(), ErrFormatInvalidRenderer)
}

func (e *engine) template(ctx context.Context, c Compiler, format Format, ref loader.Ref) (Template, error) {
	if format.IsZero() {
		return nil, wrapErr("load", format, "", ErrEmptyRef)
	}

	key := ref.String() + ":" + format.String()

	if t, ok := e.cache.Get(ctx, key); ok {
		return t, nil
	}

	src, err := e.loader.Load(ctx, ref)
	if err != nil {
		return nil, wrapErr("load", format, ref.String(), err)
	}
	defer src.Close()

	tmpl, err := c.Compile(ctx, src, ref.String())
	if err != nil {
		return nil, wrapErr("compile", format, ref.String(), err)
	}

	_ = e.cache.Set(ctx, key, tmpl)
	return tmpl, nil
}

func (e *engine) Prewarm(ctx context.Context, f Format, refs ...loader.Ref) error {
	rd, err := e.registry.Resolve(f)
	if err != nil {
		return wrapErr("resolve", f, "", err)
	}

	c, ok := rd.(Compiler)
	if !ok {
		return nil // nothing to preheat for an Encoder
	}

	for _, ref := range refs {
		if _, err := e.template(ctx, c, f, ref); err != nil {
			return err
		}
	}

	return nil
}

func (e *engine) ContentType(f Format) (string, error) {
	rd, err := e.registry.Resolve(f)
	if err != nil {
		return "", err
	}
	return rd.ContentType(), nil
}

func (e *engine) Formates() []Format {
	return e.registry.Formats()
}
