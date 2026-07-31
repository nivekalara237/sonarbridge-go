package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimite(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(10, 20)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !limiter.Allow() {
			http.Error(writer, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
