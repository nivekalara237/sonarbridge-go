package bootstrap

import (
	"log/slog"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/infra/repository/report"
)

type App struct {
	Service *usecase.Service
	Logger  *slog.Logger
}

func NewApp(config configs.Config) *App {
	svc := &usecase.Service{
		SonarInteractor:  sonar.New(config),
		GitlabInteractor: gitlab.New(config),
		ReportInteractor: &repository.Interactor{},
		Repository: *repository.New(
			report.NewRepository(),
		),
	}

	return &App{
		Service: svc,
		// Logger:  logging.NewNoop(config),
	}
}
