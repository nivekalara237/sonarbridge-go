# Câblage du moteur

> Remplacer `github.com/example/sonarbridge-go` par le module path réel :
> `find . -name '*.go' | xargs sed -i 's|github.com/example/sonarbridge-go|<ton-module>|g'`

## Composition au démarrage

```go
package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/example/sonarbridge-go/pkg/render"
	"github.com/example/sonarbridge-go/pkg/render/cache"
	"github.com/example/sonarbridge-go/pkg/render/loader"
	"github.com/example/sonarbridge-go/pkg/render/renderers"
)

//go:embed templates/*
var templatesFS embed.FS

func newEngine() (render.Engine, error) {
	// L'ordre de la chaîne compte : String est permissif, il passe en dernier.
	src := loader.NewChain(
		loader.NewEmbed(templatesFS),
		loader.NewFile(os.Getenv("TEMPLATE_ROOT")),
		loader.NewHTTP(&http.Client{Timeout: 5 * time.Second}, 4<<20),
		loader.NewString(),
	)

	reg := render.NewRegistry()

	funcs := template.FuncMap{
		"severityIcon": severityIcon,
		"minutes":      humanizeMinutes,
	}

	// Register instancie et valide immédiatement : un renderer cassé
	// empêche le démarrage au lieu d'échouer sur la première requête.
	for f, factory := range map[render.Format]render.Factory{
		render.FormatMarkdown: func() (render.Renderer, error) { return renderers.NewMarkdown(funcs), nil },
		render.FormatHTML:     func() (render.Renderer, error) { return renderers.NewHTML(nil), nil },
		render.FormatJSON:     func() (render.Renderer, error) { return renderers.NewJSON(""), nil },
	} {
		if err := reg.Register(f, factory); err != nil {
			return nil, err
		}
	}

	return render.New(
		render.WithLoader(src),
		render.WithRegistry(reg),
		render.WithCache(cache.NewLRU(128)),
		render.WithDefaultFormat(render.Format(os.Getenv("RENDER_DEFAULT_FORMAT"))),
		render.WithAtomicWrites(true),
	)
}
```

## Préchauffage au boot

```go
ctx := context.Background()

if err := eng.Prewarm(ctx, render.FormatMarkdown,
	"embed://templates/report.md",
	"embed://templates/summary.md",
); err != nil {
	return err
}
```

Sans ça, la première requête après un scale-up paie la compilation.

## Rendu d'un rapport Sonar

```go
report := buildReport(qualityGate, measures, issues)

err := eng.Render(ctx, render.Request{
	Ref:    "embed://templates/report.md",
	Format: render.FormatMarkdown,
	Data:   report,
}, os.Stdout)
```

## Handler HTTP multi-format

Aucun switch sur le format côté appelant — `ContentType` vient du renderer.

```go
func handler(eng render.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := render.Format(r.URL.Query().Get("format"))

		ct, err := eng.ContentType(format)
		if errors.Is(err, render.ErrFormatNotRegistered) {
			http.Error(w, "unsupported format", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", ct)

		req := render.Request{
			Ref:    "embed://templates/report.md",
			Format: format,
			Data:   buildReport(...),
		}

		if err := eng.Render(r.Context(), req, w); err != nil {
			http.Error(w, "render failed", http.StatusInternalServerError)
			return
		}
	}
}
```

Note : avec `WithAtomicWrites(true)`, l'erreur remonte avant qu'un octet
ne soit écrit, donc le `http.Error` est valide. Sans le mode atomique,
le header 200 serait déjà parti et il faudrait fermer brutalement.

## Ajouter un format — le test de l'OCP

```go
// renderers/yaml.go
type YAML struct{}

var _ render.DirectRenderer = (*YAML)(nil)

func (r *YAML) Format() render.Format { return render.Format("yaml") }
func (r *YAML) ContentType() string   { return "application/yaml" }

func (r *YAML) Encode(ctx context.Context, w io.Writer, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return yaml.NewEncoder(w).Encode(data)
}
```

Puis une ligne au démarrage :

```go
reg.Register("yaml", func() (render.Renderer, error) { return &renderers.YAML{}, nil })
```

Zéro modification de `engine.go`, `registry.go` ou d'un quelconque\
appelant. C'est la propriété qu'on cherchait.

## Ajouter une source — même principe

```go
// loader/s3.go
type S3 struct{ client *s3.Client }

func (l *S3) Supports(ref loader.Ref) bool { return ref.Scheme() == "s3" }
func (l *S3) Load(ctx context.Context, ref loader.Ref) (io.ReadCloser, error) { ... }
```

Puis l'insérer dans la chaîne. `chain.go` n'est pas touché.

## Configuration (12-factor III)

| Variable | Rôle |
|---|---|
| `RENDER_DEFAULT_FORMAT` | format quand la requête n'en précise pas |
| `TEMPLATE_ROOT` | racine du loader filesystem |
| `RENDER_CACHE_SIZE` | capacité du LRU |
| `RENDER_HTTP_TIMEOUT` | timeout du loader distant |
| `RENDER_MAX_SOURCE_BYTES` | borne d'allocation par source distante |

Aucune de ces valeurs n'est en dur dans le code — le même binaire tourne
en dev, recette et prod.