package httpclient

import (
	"encoding/json"
	"net/http"
)

func IsSuccess(r *http.Response) bool {
	return r.StatusCode <= 299 && r.StatusCode >= 200
}

func ToPojo[T any](r *http.Response) *T {
	var pojo *T
	err := json.NewDecoder(r.Body).Decode(&pojo)
	if err != nil {
		return nil
	}
	return pojo
}
