package renderers

import (
	"context"
	"encoding/json"
	"io"
	"sonarbridge-go/pkg/render"
)

type JSON struct {
	indent string
}

var _ render.DirectRenderer = (*JSON)(nil)

func NewJSON(indent string) *JSON { return &JSON{indent: indent} }

func (r *JSON) Format() render.Format { return render.FormatJSON }
func (r *JSON) ContentType() string   { return "application/json; charset=utf-8" }

func (r *JSON) Encode(ctx context.Context, w io.Writer, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	if r.indent != "" {
		enc.SetIndent("", r.indent)
	}
	return enc.Encode(data)
}
