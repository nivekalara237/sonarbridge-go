package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sonarbridge-go/internal/infra/entrypoints/rest/validation"
	"sonarbridge-go/internal/logging"
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

	if apiErr.Status >= 400 {
		slog.ErrorContext(r.Context(), "request failed", "code", apiErr.Code, "status_code", apiErr.Status, "err", err, "path", r.URL.Path)
	}

	if e, ok := errors.AsType[validation.ValidationError](apiErr.err); ok {
		apiErr.SubErrors = []any{e}
	}
	if e, ok := errors.AsType[validation.ValidationErrors](apiErr.err); ok {
		for _, er := range e.Errors {
			apiErr.SubErrors = append(apiErr.SubErrors, er)
		}
	}

	WriteJSON(w, apiErr.Status, errorBody{
		Error:     apiErr,
		RequestID: w.Header().Get("X-Request-ID"),
	})
}

func GetRequestBody[T any](r *http.Request, bodyOutput *T) error {
	// var payload *T
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&bodyOutput); err != nil {
		logging.Warn("invalid payload from http.request", "error", err)
		return err
	}
	return nil
}
