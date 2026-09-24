package service

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/httpclient"
	stringify "sonarbridge-go/pkg/string"
	"sonarbridge-go/plugins/vcs/gitlab/dto"
	"strconv"

	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/comment"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/vcs/common"
)

type CommentAndNoteService struct {
	comment.UnimplementedCommentAndNoteServiceServer
	Client *httpclient.Client
}

func NewCommentAndNoteService(client *httpclient.Client) *CommentAndNoteService {
	return &CommentAndNoteService{Client: client}
}

func (s *CommentAndNoteService) CreatePullrequestCommentOrNote(ctx context.Context, req *comment.CreateCommentRequest) (*comment.CreateOrUpdateCommentResponse, error) {
	return s.createNewComment(ctx, req.ProjectId, req.PullrequestId, req.Body, "")
}
func (s *CommentAndNoteService) UpdatePullrequestComment(ctx context.Context, req *comment.UpdateCommentRequest) (*comment.CreateOrUpdateCommentResponse, error) {
	return s.createNewComment(ctx, req.ProjectId, req.PullrequestId, req.Body, req.CommentId)
}

func (s *CommentAndNoteService) DeletePullrequestComment(ctx context.Context, req *comment.DeleteCommentRequest) (*comment.DeleteCommentResponse, error) {
	httpresponse, err := s.Client.Delete(ctx, "/projects/"+utils.EncodeURIComponent(req.ProjectId)+"/merge_requests/"+req.PullrequestId+"/notes/"+req.CommentId,
		&httpclient.RequestOptions{})

	if err != nil {
		logging.Warn("échec de suppression du commentaire (commit note)", "merge_request_iid", req.PullrequestId, "error", err)
		return nil, fmt.Errorf("échec création du commentaire: [%w]", err)
	}

	if httpclient.IsSuccess(httpresponse) {
		return &comment.DeleteCommentResponse{
			Metadata: &common.HttpResponseMetadata{
				Status:   int32(httpresponse.StatusCode),
				ErrorMsg: new("Unable to parse json"),
			},
		}, nil
	}

	return nil, err
}

func (s *CommentAndNoteService) createNewComment(ctx context.Context, projectId, mergeRequestId, noteBody string, nodeId string) (*comment.CreateOrUpdateCommentResponse, error) {

	if len(noteBody) > 1_000_000 {
		return nil, fmt.Errorf("the content of a note is limited to 1,000,000 characters")
	}

	data, _ := json.Marshal(map[string]string{"body": noteBody})
	var httpResponse *http.Response
	var err error
	if nodeId == "" {
		httpResponse, err = s.Client.Post(
			ctx,
			bPath(projectId, mergeRequestId),
			&httpclient.RequestOptions{Body: data})
	} else {
		httpResponse, err = s.Client.Put(
			ctx,
			notePath(projectId, mergeRequestId, nodeId),
			&httpclient.RequestOptions{Body: data})
	}

	if err != nil {
		logging.Warn("échec de "+utils.Ternary(nodeId == "", "création", "mise à jour")+" du commentaire (commit note)", "merge_request_iid", mergeRequestId, "error", err)
		return nil, fmt.Errorf("échec création du commentaire: [%w]", err)
	}

	if httpclient.IsSuccess(httpResponse) {
		created := httpclient.ToPojo[dto.Note](httpResponse)
		if created == nil {
			return &comment.CreateOrUpdateCommentResponse{
				Metadata: &common.HttpResponseMetadata{
					Status:   int32(httpResponse.StatusCode),
					ErrorMsg: new("Unable to parse json"),
				},
			}, nil
		}

		logging.Info("created new comment on MR !", "mrIID", mergeRequestId)

		id := strconv.FormatInt(created.ID, 10)

		return &comment.CreateOrUpdateCommentResponse{
			Metadata: &common.HttpResponseMetadata{
				Status:   int32(httpResponse.StatusCode),
				ErrorMsg: nil,
			},
			Data: &comment.CommentItemResponse{
				Id:              created.ID,
				Body:            created.Body,
				CreatedAt:       created.CreatedAt,
				UpdatedAt:       created.UpdatedAt,
				System:          new(created.System),
				NoteOrCommentId: &id,
				NoteUrl:         nil,
				Author: &common.Author{
					Id:           stringify.ToString(created.Author.ID),
					Username:     created.Author.Username,
					DisplayName:  created.Author.Name,
					Email:        new(created.Author.PublicEmail),
					AvatarUrl:    new(created.Author.AvatarURL),
					ProfilWebUrl: new(created.Author.WebURL),
					State:        new(created.Author.State),
					Type:         nil,
				},
				Others: nil,
			},
		}, nil
	}

	var errText []byte
	httpResponse.Body.Read(errText)
	return &comment.CreateOrUpdateCommentResponse{
		Metadata: &common.HttpResponseMetadata{
			Status:   int32(httpResponse.StatusCode),
			ErrorMsg: new(string(errText)),
		},
	}, fmt.Errorf(string(errText))
}

func bPath(projectID, mrIID string) string {
	return fmt.Sprintf("/projects/%s/merge_requests/%s/notes", utils.EncodeURIComponent(projectID), mrIID)
}

func notePath(projectID, mrIID, noteId string) string {
	return bPath(projectID, mrIID) + "/" + noteId
}
