package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/dto"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/logging"
	"time"
)

type WebhookHandler struct {
	service *usecase.Service
}

func NewWebhookHandler(svc *usecase.Service) *WebhookHandler {
	return &WebhookHandler{
		service: svc,
	}
}

func (useCase *WebhookHandler) Handler(writer http.ResponseWriter, request *http.Request) error {
	if request.Method != http.MethodPost {
		return httpx.ErrMethodNotAllowed
	}

	var payload dto.WebhookRequestDto
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		logging.Error("payload invalide", "error", err)
		return httpx.ErrBadRequest
	}
	logging.Info("payload", payload)

	logging.Info("webhook reçu",
		"project", payload.SonarProject.Key,
		"mrIID", payload.MergeRequest.IID,
		"branch", payload.GitLab.Branch,
	)

	// r.Context() : annulé automatiquement si le client HTTP (GitLab-CI) coupe la connexion.
	// On ajoute une marge de sécurité de 15s pour les appels sortants vers Sonar.
	ctx, cancel := context.WithTimeout(request.Context(), 15*time.Second)
	defer cancel()

	response, err := useCase.service.Execute(ctx, domain.SonarQubeWebhookPayload{
		TaskID: &payload.TaskID,
		Status: "",
		SonarProject: domain.Project{
			Key: payload.SonarProject.Key,
		},
		GitLab: domain.GitLab{
			ProjectID: payload.GitLab.ProjectID,
			CIToken:   payload.GitLab.CIToken,
		},
		Branch: &domain.Branch{
			Name: payload.GitLab.Branch,
			URL:  payload.GitLab.BranchUrl,
			Commit: &domain.Commit{
				SHA:     payload.GitLab.CommitSha,
				Message: "",
			},
		},
		MergeRequest: &domain.MergeRequest{
			IID: payload.MergeRequest.IID,
		},
		Properties: nil,
	})
	if err != nil {
		slog.Error("erreur inattendue", err)
		// http.Error(writer, err.Error(), http.StatusInternalServerError)
		return httpx.ErrInternal
	}

	if !response.Mergeable {
		writer.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	writer.Header().Set("Content-Type", "application/json")
	// fmt.Fprintf(writer, `{"received":%t,"qualityGateStatus":"%s","mergeable":%t}`, response.Received, response.QualityGateStatus, response.Mergeable)
	httpx.WriteJSON(writer, 200, &response)
	return nil
}
