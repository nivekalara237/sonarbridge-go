package rest

import (
	"net/http"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
)

type ReportHandler struct {
	reportService *usecase.ReportService
}

func NewReportHandler(svc *usecase.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: svc,
	}
}

func (h *ReportHandler) Handler(w http.ResponseWriter, request *http.Request) error {
	id := request.PathValue("id")
	_ = request.PathValue("output")

	report, err := h.reportService.ExecuteGet(request.Context(), id)
	if err != nil {
		return httpx.New(http.StatusNotFound, "not_found", "report with id %s is not found or something wrong white getting it", id)
	}
	httpx.WriteJSON(w, http.StatusOK, report)

	return nil
}
