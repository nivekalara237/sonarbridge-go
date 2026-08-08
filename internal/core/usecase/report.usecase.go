package usecase

import (
	"context"
	"fmt"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/logging"
)

type ReportService struct {
	ReportInteractor interactor.ReportInteractorV2
}

func NewReportService(interactor interactor.ReportInteractorV2) *ReportService {
	return &ReportService{ReportInteractor: interactor}
}

func (s *ReportService) ExecuteSave(ctx context.Context, report *sonar.Report) error {
	if report != nil {
		logging.Error("trying to save nil report")
		return fmt.Errorf("report cannot be nit")
	}
	return s.ReportInteractor.SaveReport(ctx, report)
}

func (s *ReportService) ExecuteGet(ctx context.Context, id string) (*sonar.Report, error) {
	if id == "" {
		logging.Error("report id cannot be empty")
		return nil, fmt.Errorf("report id cannot be empty")
	}
	return s.ReportInteractor.GetReport(ctx, id)
}
