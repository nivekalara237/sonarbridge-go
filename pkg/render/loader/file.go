package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
)

// ******* File *******

// File load form filesystem, under an imposed root
type File struct {
	root string
}

func NewFile(root string) *File {
	return &File{root: root}
}

func (l *File) Supports(ref Ref) bool {
	return ref.Scheme() == "file"
}

func (l *File) Load(_ context.Context, ref Ref) (io.ReadCloser, error) {
	f, err := os.Open(l.resolve(ref))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, ref)
	}

	if err != nil {
		return nil, err
	}

	return f, nil
}

func (l *File) Stamp(_ context.Context, ref Ref) (Stamp, error) {
	fi, err := os.Stat(l.resolve(ref))
	if os.IsNotExist(err) {
		return Stamp{}, fmt.Errorf("%w: %s", ErrNotFound, err)
	}

	if err != nil {
		return Stamp{}, err
	}
	return Stamp{ModTime: fi.ModTime(), Size: fi.Size()}, nil
}

// resolve clean the path and confine under root. without this
// "file://../../etc/passwd" can go out the root
func (l *File) resolve(ref Ref) string {
	return path.Join(l.root, path.Clean("/"+ref.Path()))
}
