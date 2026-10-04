package grpc

import (
	"context"
	"fmt"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/plugin/manager"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/capability"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/comment"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/common"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewVcs(mgr *manager.Manager) *VcsClient {
	return &VcsClient{m: mgr}
}

func (c *VcsClient) UpdatePullrequestComment(ctx context.Context, prComment domain.PullrequestComment, nodeId, provider string) (*domain.PullrequestCommentItem, error) {
	return nil, nil
}

func (c *VcsClient) CreatePullrequestComment(ctx context.Context, prComment domain.PullrequestComment, provider string) (*domain.PullrequestCommentItem, error) {
	client, err := c.commentAndNoteClient(provider, capability.PullRequestCreateComment)
	if err != nil {
		return nil, err
	}

	extra := make(map[string]*structpb.Value, len(prComment.Others))
	if len(prComment.Others) > 0 {
		for _, ex := range prComment.Others {
			v, er := structpb.NewValue(ex.Value)
			if er != nil {
				continue
			}
			extra[ex.Key] = v
		}
	}

	resp, err := client.CreatePullrequestCommentOrNote(ctx, &comment.CreateCommentRequest{
		ProjectId:     prComment.ProjectId,
		PullrequestId: prComment.PullrequestId,
		Body:          prComment.CommentBody,
		Others:        &common.ExtendedData{Data: &structpb.Struct{Fields: extra}},
	})

	if err != nil {
		return nil, fmt.Errorf("vcs: %s: create comment: %w", provider, err)
	}

	// debug.PrintLn(resp.Data, err)

	if resp.Metadata != nil && !isSuccessStatus(resp.Metadata.GetStatus()) {
		return nil, fmt.Errorf("vcs: %s: %s: upstream status %d: %s", provider, "create comment", resp.Metadata.GetStatus(), errMsg(resp.Metadata))
	}

	typeFn := func(t *common.UserType) string {
		if t == nil {
			return ""
		}
		return t.String()
	}

	strFn := func(s *string) string {
		if s != nil {
			return *s
		}
		return ""
	}

	return &domain.PullrequestCommentItem{
		CommentId:  resp.Data.Id,
		IsSystem:   *resp.Data.System,
		Body:       resp.Data.Body,
		CreatedAt:  resp.Data.CreatedAt,
		UpdatedAt:  resp.Data.UpdatedAt,
		CommentUrl: strFn(resp.Data.NoteUrl),
		Author: domain.Author{
			Id:           resp.Data.Author.Id,
			Username:     resp.Data.Author.Username,
			DisplayName:  resp.Data.Author.DisplayName,
			Email:        *resp.Data.Author.Email,
			AvatarUrl:    *resp.Data.Author.AvatarUrl,
			ProfilWebUrl: *resp.Data.Author.ProfilWebUrl,

			State: *resp.Data.Author.State,
			Type:  typeFn(resp.Data.Author.Type),
		},
		Others: domain.ExtraData{},
	}, nil
}

func (c *VcsClient) GetMergeRequest(ctx context.Context, projectId, mergeRequestId, provider string) (any, error) {
	return nil, nil
}
