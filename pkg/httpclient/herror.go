package httpclient

import (
	"fmt"
	"io"
	"net/http"
)

type ClientError struct {
	StatusCode int
	Status     string
	Message    string
	Body       []byte
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
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
