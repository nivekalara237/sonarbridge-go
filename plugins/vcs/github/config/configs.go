package config

import (
	"fmt"
	"os"
)

type GithubPluginConfig struct {
	BaseUrl    string
	Token      string
	Repo       string
	Owner      string
	CACertPath string
}

func LoadConfigs() (*GithubPluginConfig, error) {
	baseUrl := os.Getenv("GITHUB_BASE_URL")
	token := os.Getenv("GITHUB_TOKEN")
	cacertFile := os.Getenv("GITHUB_CA_CERT")
	//repo := os.Getenv("GITHUB_REPO")
	// owner := os.Getenv("GITHUB_OWNER")

	if baseUrl == "" || token == "" { // || repo == "" || owner == "" {
		return nil, fmt.Errorf("github plugin: GITHUB_BASE_URL and GITHUB_TOKEN must both be set")
	}

	return &GithubPluginConfig{
		BaseUrl:    baseUrl,
		Token:      token,
		CACertPath: cacertFile,
		// Repo:       repo,
		// Owner:      owner,
	}, nil
}
