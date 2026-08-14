package loader

import (
	"context"
	"fmt"
	"io"
	"io/fs"
)

// ***** Embed *****

// Embed load from fs.FS, typically a //go:embed
type Embed struct {
	fsys fs.FS
}

func NewEmbed(fsys fs.FS) *Embed {
	return &Embed{fsys: fsys}
}

func (l *Embed) Supports(ref Ref) bool {
	return ref.Scheme() == "embed"
}

func (l *Embed) Load(_ context.Context, ref Ref) (io.ReadCloser, error) {
	f, err := l.fsys.Open(ref.Path())
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, ref)
	}
	return f, nil
}
