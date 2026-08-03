package rest

import (
	"net/http"
	"sonarbridge-go/configs"
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

	mux.HandleFunc("GET /healthz", health.GetZ)
	mux.HandleFunc("POST /webhook/sonar", webhook.Handler)
	mux.HandleFunc("GET /projects/{id}", report.Handler)

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
		// Add().
		Build()
}
