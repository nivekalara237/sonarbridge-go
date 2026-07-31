package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
)

type GitlabInteractor interface {
	CreateCommitStatus(ctx context.Context, projectId, sha, ciToken string, statusData domain.GitlabCommitStatus) (any, error)
	CreateOrUpdateMergeRequestComment(ctx context.Context, projectId, mergeRequestId, comment, ciToken string) (*domain.GitLabNote, error)
	GetMergeRequest(ctx context.Context, projectId, mergeRequestId, ciToken string) (*domain.GitLabMergeRequest, error)
}
