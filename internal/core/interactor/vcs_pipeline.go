package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
)

type VcsPipelineInteractor interface {
	CreateCommitStatus(ctx context.Context, projectId, sha, ciToken string, statusData domain.GitlabCommitStatus) (any, error)
}
