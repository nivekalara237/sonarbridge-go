package middleware

import (
	"net/http"
	"sonarbridge-go/internal/build"
)

func HeaderAppInfo(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-App-Version", build.Version)
		next.ServeHTTP(writer, request)
	})

}
