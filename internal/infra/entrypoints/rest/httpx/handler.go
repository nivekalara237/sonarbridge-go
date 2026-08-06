package httpx

import (
	"net/http"
)

type Handlerx func(http.ResponseWriter, *http.Request) error

func (h Handlerx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		WriteError(w, r, err)
	}
}
