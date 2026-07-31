package middleware

import "net/http"

func HeaderAppInfo(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-App-Version", "1.0")
		next.ServeHTTP(writer, request)
	})

}
