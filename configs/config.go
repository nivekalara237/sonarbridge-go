package configs

import (
	"log"
	"sonarbridge-go/internal/infra/utils"
	"strconv"

	"github.com/joho/godotenv"
)

type CorsConfig struct {
	Origins          string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
}

type Config struct {
	Env  string
	Port string

	SonarBaseUrl string
	SonarToken   string

	GitlabBaseUrl string
	GitlabToken   string

	WebhookSecret string

	LogLevel string

	Cors *CorsConfig
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using environment variables")
	}

	allowCreds, _ := strconv.ParseBool(utils.GetEnvOrDefault("CORS_ALLOW_CREDENTIALS", "false"))

	return &Config{
		Env:           utils.GetEnvOrDefault("ENV", "dev"),
		Port:          utils.GetEnvOrDefault("PORT", "3045"),
		SonarBaseUrl:  utils.GetRequiredEnv("SONARQUBE_URL"),
		SonarToken:    utils.GetRequiredEnv("SONARQUBE_TOKEN"),
		GitlabBaseUrl: utils.GetRequiredEnv("GITLAB_URL"),
		GitlabToken:   utils.GetRequiredEnv("GITLAB_TOKEN"),
		WebhookSecret: utils.GetRequiredEnv("WEBHOOK_SECRET"),

		LogLevel: utils.GetEnvOrDefault("LOG_LEVEL", "info"),

		Cors: &CorsConfig{
			Origins:          utils.GetEnvOrDefault("CORS_ORIGIN", "*"),
			AllowedMethods:   utils.GetEnvOrDefault("CORS_ALLOWED_METHODS", "*"),
			AllowedHeaders:   utils.GetEnvOrDefault("CORS_ALLOWED_HEADERS", "*"),
			AllowCredentials: allowCreds,
		},
	}

}
