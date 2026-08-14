package loader

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"
)

var (
	ErrNotSupported = errors.New("loader: ref not support")
	ErrNotFound     = errors.New("loader: source not found")
)

// Ref identify the template source
// The format depend on loader: "file://tmp/x.tmpl", "https://...",
// "{{ .Name }}...{{ .end }}", "embed://templates/report.md", or from
// string
type Ref string

type AbstractLoader interface {
	// Load open the template. The caller must close the ReadClosed
	Load(ctx context.Context, ref Ref) (io.ReadCloser, error)

	// Supports indicate whether the loader known to treat this ref
	Supports(ref Ref) bool
}

// Stamper is implemented by the loader able to date a source
// This respect the SOLID principle of Segregation Interface like
// StringLoader and EmbedLoader will not implement it.
type Stamper interface {
	Stamp(ctx context.Context, ref Ref) (Stamp, error)
}

type Stamp struct {
	ETag    string    // for HTTP
	ModTime time.Time // for FileSystem
	Size    int64
}

func (r Ref) String() string {
	return string(r)
}

func (r Ref) IsZero() bool {
	return r == ""
}

func (r Ref) Scheme() string {
	i := strings.Index(r.String(), "://")
	if i < 0 {
		return ""
	}
	return r.String()[:i]
}

func (r Ref) Path() string {
	i := strings.Index(r.String(), "://")
	if i < 0 {
		return r.String()
	}

	return r.String()[i+3:]
}
