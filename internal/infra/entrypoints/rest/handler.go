package rest

import (
	"net/http"
	"sonarbridge-go/internal/core/usecase"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) SonarWebhook(w http.ResponseWriter, r *http.Request) {
	return Handler{}
}
