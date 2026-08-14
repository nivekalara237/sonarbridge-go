package render

import (
	"context"
	"io"
)

// Renderer product the output in the asked format
type Renderer interface {
	// Render write result in the w.
	// Render(ctx context.Context, w io.Writer, data any) error

	// ContentType return the MIME type products (text/html, application/json,...)
	ContentType() string

	Format() Format
}

/*
// Stamper is implemented by the loader able to date a source
// This respect the SOLID principle of Segregation Interface like
// StringLoader and EmbedLoader will not implement it.
type Stamper interface {
	Stamp(ctx context.Context, ref loader.Ref) (Stamp, error)
}

type Stamp struct {
	ETag    string    // for HTTP
	ModTime time.Time // for FileSystem
	Size    int64
}*/

type Encoder interface {
	Encode(ctx context.Context, w io.Writer, data any) error
}

// Compiler transform the raw source to executable template
type Compiler interface {
	Compile(ctx context.Context, src io.Reader, name string) (Template, error)
}

// Template is a template compiled, sure for concurrent usage
type Template interface {
	Execute(ctx context.Context, w io.Writer, data any) error
	Name() string
	Format() Format
}

/*
Here we separate Compiler and Renderer because is the health of performance.
The Compiler run only one time (cold), the Template.Execute every time
requested (hot), without this separation the caching become inefficient
*/

// Composite contracts. Serve to signature and assertions compilation
type (
	TemplateRenderer interface {
		Renderer
		Compiler
	}

	DirectRenderer interface {
		Renderer
		Encoder
	}
)
