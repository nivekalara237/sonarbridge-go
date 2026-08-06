package rest

import (
	"net/http"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	mdlw "sonarbridge-go/internal/infra/entrypoints/rest/middleware"
)

type Router struct {
}

func NewRouter(
	config configs.Config,
	health *HealthHandler,
	webhook *WebhookHandler,
	report *ReportHandler,
) *http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", httpx.Handlerx(func(w http.ResponseWriter, r *http.Request) error {
		return httpx.ErrNotFound
	}))
	mux.Handle("GET /healthz", httpx.Handlerx(health.GetZ))
	mux.Handle("POST /webhook/sonar", httpx.Handlerx(webhook.Handler))
	mux.Handle("GET /projects/{id}", httpx.Handlerx(report.Handler))

	return mdlw.NewBuilder(mux).
		Add(func(handler http.Handler) http.Handler {
			return mdlw.Recovery(true, handler)
		}).Add(
		func(handler http.Handler) http.Handler {
			return mdlw.Cors(handler, *config.Cors)
		}).
		Add(mdlw.HeaderAppInfo).
		Add(mdlw.RateLimite).
		Add(mdlw.SecurityHeadersMiddleware).
		Add(mdlw.LoggingRequestMiddleware).
		Add(mdlw.RequestID).
		// Add().
		Build()
}
