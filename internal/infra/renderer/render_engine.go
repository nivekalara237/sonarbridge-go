package renderer

import (
	"fmt"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/infra/renderer/markdown"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/pkg/render"
	"sonarbridge-go/pkg/render/cache"
	"sonarbridge-go/pkg/render/loader"
	"sonarbridge-go/pkg/render/renderers"
	"strings"
	"text/template"
)

func NewRenderEngine() (render.Engine, error) {
	sources := loader.NewChain(
		loader.NewEmbed(markdown.NewMarkdownRenderer().GetTmplFS()),
		loader.NewFile(utils.GetEnvOrDefault("TEMPLATE_ROOT", "D:\\devs")),
		loader.NewString(),
	)

	registry := render.NewRegistry()

	for f, factory := range map[render.Format]render.Factory{
		render.FormatMarkdown: func() (render.Renderer, error) {
			return renderers.NewMarkdown(template.FuncMap{
				"minutes":           humanizeMinutes,
				"analysisStatus":    analysisStatus,
				"qualityGateStatus": qualityGateStatus,
				"percent":           percent,
				"repeat":            repeatFunc,
			}), nil
		},
		render.FormatJSON: func() (render.Renderer, error) {
			return renderers.NewJSON(" "), nil
		},
	} {
		if err := registry.Register(f, factory); err != nil {
			return nil, err
		}
	}

	return render.New(
		render.WithLoader(sources),
		render.WithRegistry(registry),
		render.WithDefaultFormat(render.FormatMarkdown),
		render.WithCache(cache.NewLRU(128)),
	)
}

func repeatFunc(n int, word string) string {
	return strings.Repeat(word, n)
}

func humanizeMinutes(min int) string {
	if min <= 0 {
		return "0min"
	}

	d := min / (60 * 8) // Sonar compte une journée à 8h
	h := (min % (60 * 8)) / 60
	m := min % 60

	parts := make([]string, 0, 3)
	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dj", d))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dmin", m))
	}
	return strings.Join(parts, " ")
}

func qualityGateStatus(status sonar.QualityGateStatus) string {
	switch status {
	case sonar.QualityGatePassed:
		return ":green_heart: PASSED"
	case sonar.QualityGateFailed:
		return ":broken_heart: FAILED"
	default:
		return "⚠️ UNKNOWN"
	}
}
func analysisStatus(status sonar.AnalysisStatus) string {

	switch status {

	case sonar.AnalysisSuccess:
		return "✅ PASSED"

	case sonar.AnalysisFailed:
		return "❌ FAILED"

	default:
		return "⚠️ UNKNOWN"
	}
}

func percent(v *float64) string {

	if v == nil {
		return "-"
	}

	return fmt.Sprintf("%.1f%%", *v)
}
