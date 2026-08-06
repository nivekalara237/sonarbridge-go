package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorBody struct {
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = ErrInternal.Wrap(err)
	}

	if apiErr.Status >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "code", apiErr.Code, "err", err, "path", r.URL.Path)
	}

	WriteJSON(w, apiErr.Status, errorBody{
		Error:     apiErr,
		RequestID: w.Header().Get("X-Request-ID"),
	})
}
