package utils

import (
	"log/slog"
	"os"
)

func GetEnvOrDefault(key, orElse string) string {
	return getEnvOrElse(key, orElse)
}

func GetRequiredEnv(key string) string {
	return getRequired(key)
}

func getEnvOrElse(key string, _default string) string {
	v := os.Getenv(key)
	if v == "" {
		return _default
	}
	return v
}

func getRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("variable d'environnement manquante", "key", key)
		os.Exit(1)
	}
	return v
}
