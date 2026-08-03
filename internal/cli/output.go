package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

type Output struct {
	w      io.Writer
	format string
	Stderr io.Writer
}

func NewOutput(w io.Writer, format string) *Output {
	return &Output{
		w:      w,
		format: format,
	}
}

func (o *Output) Print(v any) error {
	if o.format == "json" {
		enc := json.NewEncoder(o.w)
		enc.SetIndent("", " ")
		return enc.Encode(v)
	}

	_, err := fmt.Fprintln(o.w, v)
	return err
}

func (o *Output) Printf(format string, args ...any) error {
	_, err := fmt.Fprintln(o.w, format, args)
	return err
}

func (o *Output) Error(err error) {
	fmt.Fprintln(o.Stderr, err)
}

func (o *Output) Debug(enabled bool, format string, args ...any) {
	if enabled {
		fmt.Fprintf(o.w, "[DEBUG] "+format+"\n", args...)
	}
}
