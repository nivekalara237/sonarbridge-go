package renderers

import (
	"context"
	"html/template"
	"io"
	"sonarbridge-go/pkg/render"
)

type HTML struct {
	funcs template.FuncMap
}

type htmlTemplate struct {
	tpl  *template.Template
	name string
}

var _ render.TemplateRenderer = (*HTML)(nil)

func NewHTML(funcs template.FuncMap) *HTML {
	return &HTML{funcs: funcs}
}

func (r *HTML) ContentType() string { return "text/html; charset=utf-8" }

func (r *HTML) Format() render.Format {
	return render.FormatHTML
}

func (r *HTML) Compile(_ context.Context, src io.Reader, name string) (render.Template, error) {
	raw, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	t, err := template.New(name).Funcs(r.funcs).Parse(string(raw))
	if err != nil {
		return nil, err
	}
	return &htmlTemplate{tpl: t, name: name}, nil
}

func (t *htmlTemplate) Name() string {
	return t.name
}

func (t *htmlTemplate) Format() render.Format {
	return render.FormatHTML
}

func (t *htmlTemplate) Execute(ctx context.Context, w io.Writer, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return t.tpl.Execute(w, data)
}
