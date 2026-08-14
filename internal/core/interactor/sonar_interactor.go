package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/domain/sonar"
)

type SonarInteractor interface {
	GetTaskDetails(ctx context.Context, taskId string) (*domain.SonarTaskDetails, error)

	GetAnalysisDetails(ctx context.Context, projectKey, branch, taskId string, taskStatus domain.TaskStatus) (*domain.AnalysisDetails, error)

	GetLatestAnalysis(ctx context.Context, projectKey, branch string) (*domain.AnalysisDetails, error)

	GetMeasures(ctx context.Context, projectKey, branch string) (*sonar.Measures, error)

	GetIssues(ctx context.Context, projectKey, branch string) ([]sonar.Issue, error)
}
