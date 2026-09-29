package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/infra/repository/report"
	"sonarbridge-go/internal/plugin/manager"
	registry2 "sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	state2 "sonarbridge-go/internal/plugin/state"
	"time"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/comment"
)

type App struct {
	Service *usecase.Service
	Logger  *slog.Logger
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
		// Logger:  logging.NewNoop(config),
	}
}

func (a *App) InitPlugins() error {
	root, err := os.MkdirTemp("", "sonarbridge-registry-install-*")
	pluginName := "ci-bridge-vcs-gitlab"
	if err != nil {
		return err
	}

	fmt.Println("\n faux endpoint GitLab (pour l'appel métier final)")
	mockGitLab := startMockGitLabAPI()
	defer mockGitLab.Close()

	registry := registry2.NewLocalRegistry(configs.AppConfig.Plugins.RemoteRegistry.Url)
	pluginDir := filepath.Join(root, configs.AppConfig.Plugins.Dir)
	state := state2.NewStore(filepath.Join(root, configs.AppConfig.Plugins.StateFile))
	m := manager.New(pluginDir, state, registry, func(info runtime.InstanceInfo) runtime.Adapter {
		return runtime.NewGoPluginAdapter()
	})

	fmt.Println(`Manager.Install("GitLab", "latest") - Resolve + Fetch + real checksum verification`)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := m.Install(ctx, pluginName, "latest"); err != nil {
		return fmt.Errorf("install: Gitlab Plugin: %w", err)
	}

	fmt.Println("   installed - manifest + binary write on disk, checksum verified")

	fmt.Println("enabling from boostrap (discovery what have been installed")
	if err := m.Enable(pluginName); err != nil {
		return fmt.Errorf("enable: %w", err)
	}

	if err := m.Bootstrap(); err != nil {
		return fmt.Errorf("boostrap: %w", err)
	}

	fmt.Println("Start binary")
	_ = os.Setenv("GITLAB_BASE_URL", mockGitLab.URL)
	_ = os.Setenv("GITLAB_TOKEN", "demo-token")

	defer os.Unsetenv("GITLAB_BASE_URL")
	defer os.Unsetenv("GITLAB_TOKEN")

	if err := m.Start(context.Background(), pluginName); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	for _, s := range m.List() {
		fmt.Printf("    %s -> %s (version=%s)\n", s.Name, s.State, s.Info.Version)
	}

	fmt.Println("\n[7/7] appel réel CreatePullrequestCommentOrNote sur le binaire téléchargé")
	raw, err := m.Dispense(pluginName, plugin.CommentAndNoteKey)
	if err != nil {
		return fmt.Errorf("dispense: %w", err)
	}
	commentClient := raw.(comment.CommentAndNoteServiceClient)

	resp, err := commentClient.CreatePullrequestCommentOrNote(ctx, &comment.CreateCommentRequest{
		ProjectId: "42", PullrequestId: "7", Body: "installed straight from the (fake) registry",
	})
	if err != nil {
		return fmt.Errorf("CreatePullrequestCommentOrNote: %w", err)
	}
	fmt.Printf("    status=%d note id=%d\n", resp.Metadata.Status, resp.Data.Id)

	fmt.Println("\nOK — Resolve -> Fetch -> checksum -> Install -> Bootstrap -> Start -> RPC métier, tout validé avec le vrai binaire du plugin.")
	return nil
}

func startMockGitLabAPI() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/42/merge_requests/7/notes", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 8888, "body": body["body"], "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
			"author": map[string]any{"id": 1, "username": "sonarbridge-bot"},
		})
	})
	return httptest.NewServer(mux)
}
