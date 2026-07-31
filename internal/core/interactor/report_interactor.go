package interactor

import "sonarbridge-go/internal/core/domain"

type ReportInteractor interface {
	FormatMarkdown(analysisDetails domain.AnalysisDetails) string
}
