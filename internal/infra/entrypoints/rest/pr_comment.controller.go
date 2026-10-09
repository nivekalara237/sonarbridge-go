package rest

import (
	"net/http"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/entrypoints/dto"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/infra/entrypoints/rest/validation"
)

type PullrequestCommentHandler struct {
	inter interactor.VcsPullrequestCommentOrNoteInteractor
}

func NewPullrequestComment(commentInteractor interactor.VcsPullrequestCommentOrNoteInteractor) *PullrequestCommentHandler {
	return &PullrequestCommentHandler{inter: commentInteractor}
}

func (p *PullrequestCommentHandler) UpdateCommentOrNote(
	w http.ResponseWriter,
	req *http.Request,
) error {
	return nil
}
func (p *PullrequestCommentHandler) CreateCommentOrNote(
	w http.ResponseWriter,
	req *http.Request,
) error {
	if req.Method != http.MethodPost {
		return httpx.ErrMethodNotAllowed
	}
	var createCommentRequest dto.VcsCreateCommentRequestDTO
	if err := httpx.GetRequestBody(req, &createCommentRequest); err != nil {
		return httpx.ErrBadRequest
	}

	createCommentValidation := validation.NewValidator[dto.VcsCreateCommentRequestDTO](createCommentRequest, "create_comment", nil)
	if err := createCommentValidation.IsValidOrThrow(); err != nil {
		return httpx.ErrBadRequest.Wrap(err)
	}

	var others []domain.ExtraData

	if len(createCommentRequest.Extra) > 0 {
		for _, ex := range createCommentRequest.Extra {
			others = append(others, domain.ExtraData{
				Key:   ex.Key,
				Value: ex.Value,
			})
		}
	}
	commentResponse, err := p.inter.CreatePullrequestComment(
		req.Context(),
		domain.PullrequestComment{
			ProjectId:     createCommentRequest.ProjectId,
			PullrequestId: createCommentRequest.PullrequestId,
			CommentBody:   createCommentRequest.CommentBody,
			Others:        others,
		},
		createCommentRequest.Provider,
	)
	if err != nil {
		return httpx.ErrInternal.Wrap(err)
	}
	apiresp := dto.OkResponse(commentResponse)
	httpx.WriteJSON(w, http.StatusCreated, apiresp)
	return nil
}
