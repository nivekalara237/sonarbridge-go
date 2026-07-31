package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
)

type SonarInteractor interface {
	GetTaskDetails(ctx context.Context, taskId string) (*domain.SonarTaskDetails, error)

	GetAnalysisDetails(ctx context.Context, projectKey, branch, taskId string) (*domain.AnalysisDetails, error)

	GetLatestAnalysis(ctx context.Context, projectKey, branch string) (*domain.AnalysisDetails, error)
}
