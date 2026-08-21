package httpx

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status  int    `json:"statusCode"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	err     error
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.err
}

func (e *APIError) Wrap(err error) *APIError {
	c := *e
	c.err = err
	return &c
}

func New(status int, code, msg string, args ...any) *APIError {
	m := msg
	if len(args) > 0 {
		m = fmt.Sprintf(msg, args)
	}
	return &APIError{Status: status, Code: code, Message: m}
}

var (
	ErrNotFound         = New(404, "not_found", "resource introuvable")
	ErrBadRequest       = New(400, "bad_request", "requête invalide")
	ErrUnauthorized     = New(401, "unauthorized", "authentification requise")
	ErrInternal         = New(500, "internal_error", "erreur interne")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "method_not_allowed", "méthode non autorisée")
	ErrForbidden        = New(http.StatusForbidden, "forbidden", "operation interdite")
)
