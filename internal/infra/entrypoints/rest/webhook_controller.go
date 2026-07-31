package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/dto"
	"time"
)

type UseCase struct {
	WebhookUseCase *usecase.Service
}

func (useCase *UseCase) Handler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var payload dto.WebhookRequestDto
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		slog.Error("payload invalide", "error", err)
		http.Error(writer, "payload JSON invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("webhook reçu",
		"project", payload.SonarProject.Key,
		// "status", payload.Status,
		"mrIID", payload.MergeRequest.IID,
		"branch", payload.GitLab.Branch,
	)

	// r.Context() : annulé automatiquement si le client HTTP (GitLab-CI) coupe la connexion.
	// On ajoute une marge de sécurité de 15s pour les appels sortants vers Sonar.
	ctx, cancel := context.WithTimeout(request.Context(), 15*time.Second)
	defer cancel()

	response, err := useCase.WebhookUseCase.Execute(ctx, domain.SonarQubeWebhookPayload{
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
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if !response.Mergeable {
		writer.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, `{"received":%t,"qualityGateStatus":"%s","mergeable":%t}`, response.Received, response.QualityGateStatus, response.Mergeable)
}
