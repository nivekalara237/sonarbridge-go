package httpclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ClientError struct {
	StatusCode int
	Status     string
	Message    string
	Body       []byte
	URL        string
	Method     string
	RequestID  string
}

func (e *ClientError) Error() string {
	msg := fmt.Sprintf("[HTTP %d: %s] %s %s",
		e.StatusCode, http.StatusText(e.StatusCode), e.Method, e.URL)
	if len(e.Body) > 0 {
		msg += " — " + string(e.Body)
	}
	if e.RequestID != "" {
		msg += " (request_id=" + e.RequestID + ")"
	}
	return msg
}

// Helper to check the status and return an error
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return &ClientError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       body,
		}
	}
	return nil
}

func (e *ClientError) is5xx() bool {
	return e.StatusCode <= 500 && e.StatusCode >= 599
}

func (e *ClientError) isBadRequest() bool {
	return e.StatusCode == 400
}
func (e *ClientError) isNotfound() bool {
	return e.StatusCode == 404
}

func (e *ClientError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

func IsHttpError(err error, status int) bool {
	var he *ClientError
	return errors.As(err, &he) && he.StatusCode == status
}

func (e *ClientError) Retryable() bool {
	switch e.StatusCode {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}
