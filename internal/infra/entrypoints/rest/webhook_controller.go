package rest

import (
	"context"
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

	if err := httpx.GetRequestBody[dto.WebhookRequestDto](request, &payload); err != nil {
		return httpx.ErrBadRequest
	}

	logging.Info("webhook reçu",
		"Sonar Project Key", payload.SonarProject.Key,
		"GitlabMergeRequestIID", payload.MergeRequest.IID,
		"BranchName", payload.GitLab.Branch,
		"SonarTaskId", payload.TaskID,
	)

	// r.Context() : annulé automatiquement si le client HTTP (GitLab-CI) coupe la connexion.
	// On ajoute une marge de sécurité de 15s pour les appels sortants vers Sonar.
	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()

	// reqID := middleware.RequestIDFrom(ctx)

	response, err := useCase.service.Execute(ctx, domain.SonarQubeWebhookPayload{
		TaskID: &payload.TaskID,
		Status: domain.WebhookStatus(payload.TaskStatus),
		SonarProject: domain.Project{
			Key: payload.SonarProject.Key,
		},
		GitLab: domain.GitLab{
			ProjectID: payload.GitLab.ProjectID,
			CIToken: func() string {
				if payload.GitLab.CIToken == nil {
					return ""
				}
				return *payload.GitLab.CIToken
			}(),
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
		logging.Error("erreur inattendue", err)
		return httpx.New(http.StatusInternalServerError, "Internal_error", err.Error())
	}
	code := http.StatusOK
	if !response.Mergeable {
		code = http.StatusUnprocessableEntity
	}
	// writer.Header().Set("Content-Type", "application/json")
	// fmt.Fprintf(writer, `{"received":%t,"qualityGateStatus":"%s","mergeable":%t}`, response.Received, response.QualityGateStatus, response.Mergeable)
	httpx.WriteJSON(writer, code, &response)
	return nil
}
