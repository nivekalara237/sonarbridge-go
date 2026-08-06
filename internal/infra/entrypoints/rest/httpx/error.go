package httpx

import "net/http"

type APIError struct {
	Status  int    `json:"-"`
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

func New(status int, code, msg string) *APIError {
	return &APIError{Status: status, Code: code, Message: msg}
}

var (
	ErrNotFound         = New(404, "not_found", "resource introuvable")
	ErrBadRequest       = New(400, "bad_request", "requête invalide")
	ErrUnauthorized     = New(401, "unauthorized", "authentification requise")
	ErrInternal         = New(500, "internal_error", "erreur interne")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "method_not_allowed", "méthode non autorisée")
)
