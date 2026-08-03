package middleware

import "net/http"

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-XSS-Protection", "1; mode=block")
		writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(writer, request)
	})
}
