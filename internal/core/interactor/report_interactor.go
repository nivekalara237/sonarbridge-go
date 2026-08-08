package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/domain/sonar"
)

type ReportInteractor interface {
	FormatMarkdown(analysisDetails domain.AnalysisDetails) string
}

type ReportInteractorV2 interface {
	SaveReport(ctx context.Context, report *sonar.Report) error
	GetReport(ctx context.Context, id string) (*sonar.Report, error)
}
