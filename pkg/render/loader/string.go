package loader

import (
	"context"
	"io"
	"strings"
)

// ******** String ********

// String handle the ref itself as tempate content
// Well for tests and injected tempate by configuration/cli of
// coming from data base table
type String struct{}

func NewString() *String {
	return &String{}
}

// Supports here accept every Ref without schema
func (l *String) Supports(ref Ref) bool {
	return !ref.IsZero() && ref.String() == ""
}

func (l *String) Load(_ context.Context, ref Ref) (io.ReadCloser, error) {
	if ref.IsZero() {
		return nil, ErrNotFound
	}
	return io.NopCloser(strings.NewReader(ref.String())), nil
}
