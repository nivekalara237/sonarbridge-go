package bootstrap

import (
	"log/slog"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/logging"
	"strings"
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
	}

	level := slog.LevelDebug

	switch strings.ToLower(config.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	default:
		level = slog.LevelDebug
	}

	return &App{
		Service: svc,
		Logger:  logging.New(level),
	}
}
