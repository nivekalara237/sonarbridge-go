package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/logging"
	"strconv"
	"strings"
	"time"
)

func signRequest(method, path, body string, timestamp int64, sharedSecret string) string {
	mac := hmac.New(sha256.New, []byte(sharedSecret))
	minifiedBody, err := minifyJson(body)
	if err != nil {
		return ""
	}
	payload := fmt.Sprintf("%s\n%s\n%s\n%d", method, path, minifiedBody, timestamp)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func minifyJson(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func verifySignature(secret string, r *http.Request) bool {
	sigHeader := r.Header.Get("X-Sonar-Webhook-Hmac-Sign")
	if sigHeader == "" {
		logging.Warn("Missing webhook signature")
		return false
	}
	tsHeader := r.Header.Get("X-Sonar-Webhook-Timestamp")
	if tsHeader == "" {
		logging.Warn("Missing timestamp signature")
		return false
	}

	timestamp, errt := strconv.ParseInt(tsHeader, 10, 64)
	if errt != nil {
		return false
	}

	if time.Now().UTC().Sub(time.Unix(timestamp, 0)) > 5*time.Minute {
		return false
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	expectedSig := signRequest(r.Method, r.URL.Path, string(bodyBytes), timestamp, secret)
	return hmac.Equal([]byte(sigHeader), []byte(expectedSig))
}

func AuthHmacSignature(sharedSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !unAuthenticableRoute(r) && !verifySignature(sharedSecret, r) {
			httpx.WriteError(w, r, httpx.ErrForbidden)
			logging.Error("Forbidden: invalid signature")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unAuthenticableRoute(r *http.Request) bool {
	path := r.URL.Path
	insecuredRoutes := []string{"/", "/healthz", "/docs/(.*)"}
	return slices.Contains(insecuredRoutes, path)
}
