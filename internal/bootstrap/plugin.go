package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/provision"
	"sonarbridge-go/internal/plugin/registry"
	"sonarbridge-go/internal/plugin/runtime"
	"sonarbridge-go/internal/plugin/state"
	"strings"
	"time"
)

func (a *App) initPluginManager() error {
	rootDir := os.ExpandEnv(configs.AppConfig.Plugins.Dir)
	fmt.Println("Plugin manager", ":", "Initialization", rootDir)
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return fmt.Errorf("unable to create Plugin directory: %w", err)
	}

	vcsProviders := configs.AppConfig.VcsProviders
	if len(vcsProviders) == 0 {
		return fmt.Errorf("at least one provider have to provider")
	}

	registryClient := getRegistryClient()
	if registryClient == nil {
		return fmt.Errorf("remote-registry is enabled but, unable to create registry client")
	}

	newStore := state.NewStore(resolvePath(rootDir, os.ExpandEnv(configs.AppConfig.Plugins.StateFile)))

	m := manager.New(rootDir, newStore, registryClient, func(info runtime.InstanceInfo) runtime.Adapter {
		return runtime.NewGoPluginAdapter()
	})
	a.PluginManager = m

	if err := m.Bootstrap(); err != nil {
		return fmt.Errorf("boostrap: %w", err)
	}

	return nil
}

func loadPluginEnvByVcs(vcs []configs.VcsProviderCnf) map[string]map[string]string {
	pEnvMap := make(map[string]map[string]string)
	for _, provider := range vcs {
		pEnvMap[provider.BinaryName] = map[string]string{
			fmt.Sprintf("%s_TOKEN", strings.ToUpper(provider.Name)):    provider.TokenEnvVar,
			fmt.Sprintf("%s_BASE_URL", strings.ToUpper(provider.Name)): provider.BaseUrl,
			fmt.Sprintf("%s_VERSION", strings.ToUpper(provider.Name)):  provider.Version,
			fmt.Sprintf("%s_CA_FILE", strings.ToUpper(provider.Name)):  provider.CacertFile,
		}

		if slices.Contains([]string{"github", "gitea"}, strings.ToLower(provider.Name)) {
			pEnvMap[provider.BinaryName][fmt.Sprintf("%s_OWNER", strings.ToUpper(provider.Name))] = provider.Owner
			pEnvMap[provider.BinaryName][fmt.Sprintf("%s_REPO", strings.ToUpper(provider.Name))] = provider.Repo
		}
	}
	return pEnvMap
}

func (a *App) RunPluginProvisioner(provisioned chan struct{}, pCtx context.Context) error {
	cnf := configs.AppConfig
	rootDir := os.ExpandEnv(configs.AppConfig.Plugins.Dir)

	if a.PluginManager == nil {
		if err := a.initPluginManager(); err != nil {
			a.Logger.Error("plugin provisioner", "error", err)
			return err
		}
	}

	a.PluginManager.PluginEnv = make(map[string]map[string]string, len(cnf.VcsProviders))
	toProvision := make([]provision.Desired, 0, len(cnf.VcsProviders))
	mapEnvs := loadPluginEnvByVcs(cnf.VcsProviders)
	for _, p := range cnf.VcsProviders {
		env := mapEnvs[p.BinaryName]
		toProvision = append(toProvision, provision.Desired{
			Name:    p.BinaryName,
			Version: utils.Ternary[string](p.Version == "", "latest", p.Version),
		})
		a.PluginManager.PluginEnv[p.BinaryName] = env
	}

	go func() {
		defer close(provisioned)

		prov := &provision.Provisioner{
			PluginDir:      rootDir,
			Manager:        a.PluginManager,
			InstallTimeout: 5 * time.Second,
			Log:            a.Logger,
		}

		if orphans := prov.Orphans(toProvision); len(orphans) > 0 {
			a.Logger.Warn("plugins installed but absent from the configuration are left untouched", "plugins", orphans)
		}

		failed := 0
		for _, r := range prov.Run(pCtx, toProvision) {
			if r.Err != nil {
				failed++
			}
		}

		if failed > 0 || len(toProvision) < len(cnf.VcsProviders) {
			a.Logger.Warn("server is degraded: /readyz stays false until every configured provider is READY", "configured", len(cnf.VcsProviders), "failed", failed+len(cnf.VcsProviders)-len(toProvision))
		} else {
			a.Logger.Info("plugins provisioned", "count", len(toProvision))
		}
	}()

	return nil
}

func getRegistryClient() registry.Client {
	if configs.AppConfig.Plugins.RemoteRegistry.Enabled {
		var opts []registry.LocalRegistryOption
		if configs.AppConfig.Plugins.RemoteRegistry.Username != "" ||
			configs.AppConfig.Plugins.RemoteRegistry.Password != "" {
			opts = append(opts, registry.WithBasicAuth(
				configs.AppConfig.Plugins.RemoteRegistry.Username,
				configs.AppConfig.Plugins.RemoteRegistry.Password,
			),
			)
		}
		if configs.AppConfig.Plugins.RemoteRegistry.Type == "http" {
			return registry.NewLocalRegistry(
				configs.AppConfig.Plugins.RemoteRegistry.Url,
				utils.Ternary(len(opts) > 0, opts, []registry.LocalRegistryOption{})...,
			)
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
	} else {
		return registry.NewDisableRegistry()
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
