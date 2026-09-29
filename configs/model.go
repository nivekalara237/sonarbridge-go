package configs

type AppUntypedConfig struct {
	HomeDir     string `json:"home_dir" yaml:"home_dir" mapstructure:"home_dir"`
	DisplayName string `json:"display_name" yaml:"display_name" mapstructure:"display_name"`
	Timezone    string `json:"timezone" yaml:"timezone" mapstructure:"timezone"`

	Server       ServerCnf        `json:"server" yaml:"server" mapstructure:"server"`
	Sonar        SonarCnf         `json:"sonar" yaml:"sonar" mapstructure:"sonar"`
	VcsProviders []VcsProviderCnf `json:"vcs_providers" yaml:"vcs_providers" mapstructure:"vcs_providers"`
	Logging      LoggingCnf       `json:"logging" yaml:"logging" mapstructure:"logging"`
	Plugins      PluginsCnf       `json:"plugins" yaml:"plugins" mapstructure:"plugins"`
}

type ServerCnf struct {
	Host                  string  `json:"host" yaml:"host" mapstructure:"host"`
	Port                  int     `json:"port" yaml:"port" mapstructure:"port"`
	SharedSecretKeyEnvVar string  `json:"shared_secret_key_env_var" yaml:"shared_secret_key_env_var" mapstructure:"shared_secret_key_env_var"`
	Tls                   TlsCnf  `json:"tls" yaml:"tls" mapstructure:"tls"`
	Cors                  CorsCnf `json:"cors" yaml:"cors" mapstructure:"cors"`
}

type TlsCnf struct {
	Enabled  bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file" mapstructure:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file" mapstructure:"key_file"`
	CaFile   string `json:"ca_file" yaml:"ca_file" mapstructure:"ca_file"`
}

type CorsCnf struct {
	Enabled          bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Origins          []string `json:"origins" yaml:"origins" mapstructure:"origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers" mapstructure:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAgeSeconds    int64    `json:"max_age_seconds" yaml:"max_age_seconds" mapstructure:"max_age_seconds"`
}

type SonarCnf struct {
	BaseUrl     string `json:"base_url" yaml:"base_url" mapstructure:"base_url"`
	TokenEnvVar string `json:"token_env_var" yaml:"token_env_var" mapstructure:"token_env_var"`
	CaCertFile  string `json:"ca_cert_file" yaml:"ca_cert_file" mapstructure:"ca_cert_file"`
	Edition     string `json:"edition" yaml:"edition" mapstructure:"edition"`
}

type VcsProviderCnf struct {
	BaseUrl     string `json:"base_url" yaml:"base_url" mapstructure:"base_url"`
	TokenEnvVar string `json:"token_env_var" yaml:"token_env_var" mapstructure:"token_env_var"`
	Name        string `json:"name" yaml:"name" mapstructure:"name"`
	Owner       string `json:"owner" yaml:"owner" mapstructure:"owner"`
	Repo        string `json:"repo" yaml:"repo" mapstructure:"repo"`
	CacertFile  string `json:"cacert_file" yaml:"cacert_file" mapstructure:"cacert_file"`
}

type LoggingCnf struct {
	Level  string `json:"level" yaml:"level" mapstructure:"level"`
	Format string `json:"format" yaml:"format" mapstructure:"format"`
}

type PluginsCnf struct {
	Dir               string                  `json:"dir" yaml:"dir" mapstructure:"dir"`
	StateFile         string                  `json:"state_file" yaml:"state_file" mapstructure:"state_file"`
	LocalRegistryFile string                  `json:"local_registry_file" yaml:"local_registry_file" mapstructure:"local_registry_file"`
	RemoteRegistry    PluginRemoteRegistryCnf `json:"remote_registry" yaml:"remote_registry" mapstructure:"remote_registry"`
}
type PluginRemoteRegistryCnf struct {
	Enabled     bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Type        string `json:"type" yaml:"type" mapstructure:"type"`
	RepoPrefix  string `json:"repo_prefix" yaml:"repo_prefix" mapstructure:"repo_prefix"`
	TokenEnvVar string `json:"token_env_var" yaml:"token_env_var" mapstructure:"token_env_var"`
	Username    string `json:"username" yaml:"username" mapstructure:"username"`
	Password    string `json:"password" yaml:"password" mapstructure:"password"`
	Url         string `json:"url" yaml:"url" mapstructure:"url"`
	Tls         TlsCnf `json:"tls" yaml:"tls" mapstructure:"tls"`
}

type TelemetryCnf struct {
	Enabled bool `json:"enabled"`

}
