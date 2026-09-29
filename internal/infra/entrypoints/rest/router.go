package rest

import (
	"net/http"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	mdlw "sonarbridge-go/internal/infra/entrypoints/rest/middleware"
)

type Router struct {
}

func NewRouter(
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
	mux.Handle("GET /report/{id}", httpx.Handlerx(report.Handler))

	return mdlw.NewBuilder(mux).
		Add(func(handler http.Handler) http.Handler {
			return mdlw.Recovery(true, handler)
		}).Add(
		func(handler http.Handler) http.Handler {
			if configs.AppConfig.Server.Cors.Enabled {
				return mdlw.Cors(handler, configs.AppConfig.Server.Cors)
			}
			return handler
		}).
		Add(mdlw.HeaderAppInfo).
		Add(mdlw.RateLimite).
		Add(mdlw.SecurityHeadersMiddleware).
		Add(mdlw.LoggingRequestMiddleware).
		Add(mdlw.RequestID).
		Add(func(handler http.Handler) http.Handler {
			return mdlw.AuthHmacSignature(os.Getenv(configs.AppConfig.Server.SharedSecretKeyEnvVar), handler)
		}).
		Build()
}
