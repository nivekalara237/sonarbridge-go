package bootstrap

import (
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/rest"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
)

type App struct {
	UseCase *rest.UseCase
}

func NewApp(config configs.Config) *App {
	svc := &rest.UseCase{
		WebhookUseCase: &usecase.Service{
			SonarInteractor:  sonar.New(config),
			GitlabInteractor: gitlab.New(config),
			ReportInteractor: &repository.Interactor{},
		},
	}

	return &App{
		UseCase: svc,
	}
}
