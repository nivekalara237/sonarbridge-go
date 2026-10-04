package interactor

import (
	"context"
	"sonarbridge-go/internal/core/domain"
)

type VcsPullrequestCommentOrNoteInteractor interface {
	CreatePullrequestComment(ctx context.Context, prComment domain.PullrequestComment, provider string) (*domain.PullrequestCommentItem, error)
	UpdatePullrequestComment(ctx context.Context, prComment domain.PullrequestComment, noteId, provider string) (*domain.PullrequestCommentItem, error)
}

type VcsPullrequestInteractor interface {
	GetMergeRequest(ctx context.Context, projectId, mergeRequestId, provider string) (any, error)
}
