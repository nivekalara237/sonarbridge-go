package config

import (
	"fmt"
	"os"
)

type GitlabPluginConfig struct {
	BaseUrl    string
	Token      string
	CACertPath string
}

func LoadConfigs() (*GitlabPluginConfig, error) {
	baseUrl := os.Getenv("GITLAB_BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")
	cacertFile := os.Getenv("GITLAB_CA_CERT")

	if baseUrl == "" || token == "" {
		return nil, fmt.Errorf("gitlab plugin: GITLAB_BASE_URL and GITLAB_TOKEN must both be set")
	}

	return &GitlabPluginConfig{
		BaseUrl:    baseUrl,
		Token:      token,
		CACertPath: cacertFile,
	}, nil
}
