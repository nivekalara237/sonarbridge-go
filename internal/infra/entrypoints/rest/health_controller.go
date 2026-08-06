package rest

import "net/http"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) GetZ(w http.ResponseWriter, _ *http.Request) error {
	JSON(w, 200, map[string]string{
		"status": "UP",
	})
	return nil
}
