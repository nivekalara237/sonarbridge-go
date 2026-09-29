package middleware

import (
	"net/http"
	"slices"
	"sonarbridge-go/configs"
	"strconv"
	"strings"
)

func Cors(next http.Handler, config configs.CorsCnf) http.Handler {

	allowedOrigins := make(map[string]struct{})
	for _, origin := range config.Origins {
		allowedOrigins[strings.TrimSpace(origin)] = struct{}{}
	}

	isStartOrigin := slices.ContainsFunc(config.Origins, func(o string) bool {
		return strings.TrimSpace(o) == "*"
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if _, ok := allowedOrigins[origin]; ok || isStartOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(config.AllowCredentials))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		if config.MaxAgeSeconds > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(config.MaxAgeSeconds, 10))
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
