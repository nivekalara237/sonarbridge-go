package configs

import (
	"errors"
	"fmt"
	"sonarbridge-go/internal/build"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	Origins          string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
}

type SonarQubeEdition string
type GitlabEdition string

type Config struct {
	Env  string
	Port string
	Host string

	SonarBaseUrl string
	SonarCACert  string
	SonarToken   string

	RegistryBaseUrl string

	GitlabBaseUrl string
	GitlabToken   string
	GitlabCACert  string

	WebhookSecret string

	LogLevel string

	Cors *CorsConfig

	SonarEdition  SonarQubeEdition
	GitLabEdition GitlabEdition
}

const (
	AppHomeDir = ".bridge"
	AppUser    = "bridge"
	AppName    = "Bridge CI"
)

const (
	SONARQUBE_CE        SonarQubeEdition = "community"
	SONARQUBE_DEVELOPER SonarQubeEdition = "developer"
	SONARQUBE_EE        SonarQubeEdition = "enterprise"
)

const (
	GITLAB_CE       GitlabEdition = "community"
	GITLAB_PREMIUM  GitlabEdition = "community"
	GITLAB_ULTIMATE GitlabEdition = "community"
)

var (
	Vconfig   = viper.New()
	AppConfig AppUntypedConfig
)

func InitConfig() {
	Vconfig.SetConfigName("config")
	// Vconfig.SetConfigType("toml")
	Vconfig.AddConfigPath("/etc/" + build.GetBuildInfo().Name + "/")
	Vconfig.AddConfigPath("$HOME/" + AppHomeDir)
	Vconfig.AddConfigPath(".")
	Vconfig.SetEnvPrefix(strings.ToUpper(AppUser) + "_CI")
	Vconfig.SetEnvKeyReplacer(
		strings.NewReplacer(".", "_"),
	)
	Vconfig.AutomaticEnv()
}
func LoadAppConfig() (AppUntypedConfig, error) {

	serverEnvs := []string{
		"host",
		"port",
		"cors.enabled",
		"cors.allowed_origin",
		"tls.enabled",
		"tls.cert_file",
		"tls.key_file",
	}

	for _, s := range serverEnvs {
		replace := strings.ReplaceAll(s, ".", "_")
		_ = Vconfig.BindEnv(fmt.Sprintf("server.%s", s), fmt.Sprintf("%s_CI_%s",
			strings.ToUpper(AppUser), strings.ToUpper(replace)))
	}

	var cfg AppUntypedConfig

	if err := Vconfig.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("decode configuration: maybe malformatted: %w", err)
	}

	AppConfig = cfg

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return cfg, fmt.Errorf("invalid port: %d", cfg.Server.Port)
	}

	if cfg.Server.Tls.Enabled {
		if cfg.Server.Tls.CertFile == "" {
			return cfg, fmt.Errorf("TLS enabled but certificate file is missing")
		}

		if cfg.Server.Tls.KeyFile == "" {
			return cfg, fmt.Errorf("TLS enabled but key file is missing")
		}
	}

	return cfg, nil
}

func LoadConfigFile() error {
	cfgFile := Vconfig.GetString("config")

	if cfgFile != "" {
		Vconfig.SetConfigFile(cfgFile)
		if err := Vconfig.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		return nil
	}

	if err := Vconfig.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
			return nil
		}

		return fmt.Errorf("error reading config: %w", err)
	}

	return nil
}

func SetDefaults() {
	Vconfig.SetDefault("server.port", "3045")
	Vconfig.SetDefault("server.host", "127.0.0.1")
	Vconfig.SetDefault("display_name", strings.ToTitle(AppName))
	Vconfig.SetDefault("timezone", "Europe/Paris")
	Vconfig.SetDefault("server.cors.enabled", true)
	Vconfig.SetDefault("server.cors.allow_credentials", false)
	Vconfig.SetDefault("server.cors.max_age_seconds", 3600)
	Vconfig.SetDefault("server.cors.allowed_methods", []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"OPTIONS",
	})

	Vconfig.SetDefault("server.cors.allowed_headers", []string{
		"Authorization",
		"Content-Type",
	})

	Vconfig.SetDefault("server.tls.enabled", false)
	Vconfig.SetDefault("server.tls.cert_file", "")
	Vconfig.SetDefault("server.tls.key_file", "")
	Vconfig.SetDefault("server.tls.ca_file", "")

	homeDir := "$HOME/" + AppHomeDir

	Vconfig.SetDefault("home_dir", homeDir)
	Vconfig.SetDefault("plugins.dir", homeDir+"/plugins")
	Vconfig.SetDefault("plugins.state_file", homeDir+"/plugins/state.json")
	Vconfig.SetDefault("plugins.local_registry_file", homeDir+"/plugins/registry.json")
	Vconfig.SetDefault("plugins.remote_registry.type", "http")
	Vconfig.SetDefault("plugins.remote_registry.type", "http")
	Vconfig.SetDefault("plugins.remote_registry.enabled", "true")
	Vconfig.SetDefault("plugins.remote_registry.url", "http://localhost:8077")
	Vconfig.SetDefault("logging.level", "info")
	Vconfig.SetDefault("logging.format", "json")
}

func ApplyServeFlags(flags *pflag.FlagSet) error {
	flags.String("host", "", "HTTP server host")
	flags.Int("port", 0, "HTTP server port")
	flags.Bool("cors", false, "Enable CORS")
	flags.StringSlice("cors-origin", nil, "CORS allowed origins")
	flags.Bool("tls", false, "Enable TLS")
	flags.String("tls-cert", "", "TLS certificate file")
	flags.String("tls-key", "", "TLS private key file")
	fn := func(argName, argLookupName string) error {
		if err := Vconfig.BindPFlag(argName, flags.Lookup(argLookupName)); err != nil {
			return err
		}
		return nil
	}

	var errs []error
	errs = append(errs, fn("server.host", "host"))
	errs = append(errs, fn("server.port", "port"))
	errs = append(errs, fn("server.cors.enabled", "cors"))
	errs = append(errs, fn("server.cors.origins", "cors-origin"))
	errs = append(errs, fn("server.tls.enabled", "tls"))
	errs = append(errs, fn("server.tls.cert_file", "tls-cert"))
	errs = append(errs, fn("server.tls.key_file", "tls-key"))
	var errMsgs []string

	for _, e := range errs {
		if e != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("%d: %w", len(errMsgs)+1, e))
		}
	}

	if len(errMsgs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errMsgs, "\n"))
}
