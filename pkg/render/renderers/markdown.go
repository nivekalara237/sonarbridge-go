package renderers

import (
	"context"
	"io"
	"sonarbridge-go/pkg/render"
	"text/template"
)

type Markdown struct {
	funcs template.FuncMap
}

type textTemplate struct {
	tmpl   *template.Template
	name   string
	format render.Format
}

var _ render.TemplateRenderer = (*Markdown)(nil)

func NewMarkdown(funcs template.FuncMap) *Markdown {
	return &Markdown{funcs: funcs}
}

func (r *Markdown) Format() render.Format {
	return render.FormatMarkdown
}

func (r *Markdown) ContentType() string {
	return "text/markdown; chartset=utf-8"
}

func (r *Markdown) Compile(_ context.Context, src io.Reader, name string) (render.Template, error) {
	raw, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	t, err := template.New(name).Funcs(r.funcs).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	return &textTemplate{tmpl: t, name: name, format: render.FormatMarkdown}, nil
}

func (t *textTemplate) Name() string {
	return t.name
}

func (t *textTemplate) Format() render.Format {
	return t.format
}

func (t *textTemplate) Execute(ctx context.Context, w io.Writer, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return t.tmpl.Execute(w, data)
}

func (r *Markdown) Render() {

}
