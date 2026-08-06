package rest

import (
	"fmt"
	"net/http"
	"sonarbridge-go/internal/core/usecase"
)

type ReportHandler struct {
	Service *usecase.Service
}

func NewReportHandler(svc *usecase.Service) *ReportHandler {
	return &ReportHandler{
		Service: svc,
	}
}

func (h *ReportHandler) Handler(w http.ResponseWriter, request *http.Request) error {
	fmt.Printf("Request to /projects/%s\n", request.PathValue("id"))
	fmt.Println("Queries(raw) : ", request.URL.RawQuery)
	fmt.Println("Queries : ", request.URL.Query())

	JSON(w, http.StatusOK, map[string]string{
		"report": "OK",
	})

	return nil
}
