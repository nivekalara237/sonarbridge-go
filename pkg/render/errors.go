package render

import (
	"errors"
	"fmt"
)

type Error struct {
	Op     string
	Format Format
	Ref    string
	Err    error
}

var (
	ErrFormatNotRegistered     = errors.New("render: format not supported")
	ErrFormatAlreadyRegistered = errors.New("render: form already registered")
	ErrFormatInvalidRenderer   = errors.New("render: renderer implements neither Compiler nor Encoder")
	ErrNoLoaderSupportsRef     = errors.New("render: no loader supports this ref")
	ErrEmptyRef                = errors.New("render: empty ref")
	ErrSourceTooLarge          = errors.New("render: source exceeds max size")
)

func (e *Error) Error() string {
	return fmt.Sprintf("errder: %s [format=%s ref=%s]: %v", e.Op, e.Format, e.Ref, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrapErr(op string, f Format, ref string, err error) error {
	if err == nil {
		return nil
	}

	return &Error{
		Op:     op,
		Format: f,
		Ref:    ref,
		Err:    err,
	}
}
