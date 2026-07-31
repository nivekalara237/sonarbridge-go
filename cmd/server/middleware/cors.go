package middleware

import (
	"net/http"
	"sonarbridge-go/configs"
	"strconv"
	"strings"
)

func Cors(next http.Handler, config configs.CorsConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var methods []string
		var headers []string
		if config.AllowedMethods != "" && config.AllowedMethods != "*" {
			for m := range strings.SplitSeq(config.AllowedMethods, ",") {
				methods = append(methods, strings.TrimSpace(m))
			}
		}
		if config.AllowedHeaders != "" && config.AllowedHeaders != "*" {
			for h := range strings.SplitSeq(config.AllowedHeaders, ",") {
				headers = append(headers, strings.TrimSpace(h))
			}
		}

		/**

		allowedOrigins := []string{
		        "https://myapp.com",
		        "https://www.myapp.com",
		        "https://admin.myapp.com",
		    }

		    origin := r.Header.Get("Origin")
		    for _, allowed := range allowedOrigins {
		        if origin == allowed {
		            (*w).Header().Set("Access-Control-Allow-Origin", origin)
		            break
		        }
		    }


		if config.Origins != "*" {
			incomingOrigin := r.Header.Get("Origin")

			for _, allowed := range  {

			}
		}

		*/

		w.Header().Set("Access-Control-Allow-Origin", config.Origins)
		w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(config.AllowCredentials))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
