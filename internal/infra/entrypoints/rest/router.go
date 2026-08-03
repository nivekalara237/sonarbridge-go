package rest

import "net/http"

type Router struct {
}

func NewRouter(
	health *HealthHandler,
	webhook *WebhookHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", health.GetZ)
	mux.HandleFunc("POST /webhook/sonar", webhook.Handler)

	return mux
}
