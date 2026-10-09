package bootstrap

import (
	"log/slog"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/infra/repository/report"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/internal/plugin/manager"
)

type App struct {
	Service       *usecase.Service
	Logger        *slog.Logger
	PluginManager *manager.Manager
}

func NewApp() *App {
	svc := &usecase.Service{
		SonarInteractor:  sonar.New(),
		GitlabInteractor: gitlab.New(),
		ReportInteractor: &repository.Interactor{},
		Repository: *repository.New(
			report.NewRepository(),
		),
	}

	return &App{
		Service: svc,
		Logger:  logging.GetLogger(),
	}
}

func (a *App) Bootstrapping() {
	logging.Info("Bootstrapping bridge application.")

	if configs.AppConfig.Plugins.Dir != "" {
		err := a.initPluginManager()
		if err != nil {
			panic(err)
		}
	}
}

func (a *App) SetPluginManager(mgr *manager.Manager) {
	a.PluginManager = mgr
}
