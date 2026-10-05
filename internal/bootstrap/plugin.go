package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"strings"
	"time"
)

func (a *App) initPluginManager() error {
	fmt.Println("Plugin manager", ":", "Initialization", configs.AppConfig.Plugins.Dir)
	rootDir := os.ExpandEnv(configs.AppConfig.Plugins.Dir)
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return fmt.Errorf("unable to create Plugin directory: %w", err)
	}

	vcsProviders := configs.AppConfig.VcsProviders
	if len(vcsProviders) == 0 {
		return fmt.Errorf("at least one provider have to provider")
	}

	registryClient := getRegistryClient()
	if registryClient == nil {
		return fmt.Errorf("unable to create registry client")
	}

	newStore := state.NewStore(resolvePath(rootDir, os.ExpandEnv(configs.AppConfig.Plugins.StateFile)))

	m := manager.New(rootDir, newStore, registryClient, func(info runtime.InstanceInfo) runtime.Adapter {
		return runtime.NewGoPluginAdapter()
	})
	a.PluginManager = m

	fmt.Println(`Manager.Install vcs plugin - Resolve + Fetch + real checksum verification`)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := m.Bootstrap(); err != nil {
		return fmt.Errorf("boostrap: %w", err)
	}

	for _, provider := range vcsProviders {
		pEnvMap := make(map[string]map[string]string)
		if provider.Name == "gitlab" {
			pEnvMap[provider.BinaryName] = map[string]string{
				"GITLAB_TOKEN":    provider.TokenEnvVar,
				"GITLAB_BASE_URL": provider.BaseUrl,
			}
		}
		if provider.Name == "github" {
			pEnvMap[provider.BinaryName] = map[string]string{
				"GITHUB_TOKEN":    provider.TokenEnvVar,
				"GITHUB_BASE_URL": provider.BaseUrl,
			}
		}

		m.PluginEnv = pEnvMap

		// fmt.Println(stringify.ToString(provider))
		if err := m.Install(ctx, strings.ToLower(strings.TrimSpace(provider.BinaryName)), "latest"); err != nil {
			return fmt.Errorf("install: %s Plugin: %w", provider.Name, err)
		}

		fmt.Println("enabling from boostrap (discovery what have been installed)")
		if err := m.Enable(provider.BinaryName); err != nil {
			return fmt.Errorf("enable: %w", err)
		}

		fmt.Println("Start binary")

		if err := m.Start(context.Background(), provider.BinaryName); err != nil {
			return fmt.Errorf("start: %w", err)
		}

		fmt.Printf("%s is started at %s", provider.Name, time.Now())
	}

	fmt.Println()

	for _, s := range m.List() {
		fmt.Printf("%s [%s] capabilities=%v\n\n", s.Name, s.State, s.Info.Capabilities)
	}

	return nil
}

func getRegistryClient() registry.Client {
	if configs.AppConfig.Plugins.RemoteRegistry.Enabled {
		if configs.AppConfig.Plugins.RemoteRegistry.Type == "http" {
			//TODO: complete by adding username and password
			return registry.NewLocalRegistry(configs.AppConfig.Plugins.RemoteRegistry.Url)
		}

		if configs.AppConfig.Plugins.RemoteRegistry.Type == "github" {
			return &registry.GithubRegistry{
				Token: os.Getenv(configs.AppConfig.Plugins.RemoteRegistry.TokenEnvVar),
				Owner: fmt.Sprintf("%s/%s",
					configs.AppConfig.Plugins.RemoteRegistry.GithubOwner,
					configs.AppConfig.Plugins.RemoteRegistry.GithubRepo,
				),
			}
		}
	}

	return nil
}

func resolvePath(baseDir, path string) string {
	baseDir = filepath.Clean(baseDir)
	path = filepath.Clean(path)

	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, filepath.Base(baseDir)+string(filepath.Separator)) {
		return filepath.Join(filepath.Dir(baseDir), path)
	}

	return filepath.Join(baseDir, path)
}
