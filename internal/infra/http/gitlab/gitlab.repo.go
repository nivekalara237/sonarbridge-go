package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/http"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/logging"
	"strconv"
	"strings"
)

type Interactor struct {
	interactor.GitlabInteractor
	config configs.Config
}

func New(cfg configs.Config) *Interactor {
	return &Interactor{
		GitlabInteractor: (*Interactor)(nil),
		config:           cfg,
	}
}

func (inter *Interactor) createHttpClient(ciToken string) (*http.Client, error) {
	baseUrl := inter.config.GitlabBaseUrl
	token := inter.config.GitlabToken
	if baseUrl == "" || token == "" {
		logging.Error("variable d'environnement GITLAB_API_URL ou GITLAB_TOKEN manquante")
		return nil, errors.New("variable d'environnement GITLAB_API_URL ou GITLAB_TOKEN manquante")
	}
	if strings.TrimSpace(ciToken) != "" {
		token = "ci;" + token
	}
	return http.NewClientHttp(strings.TrimRight(baseUrl, "/"), token, "gitlab"), nil
}

func (inter *Interactor) CreateCommitStatus(ctx context.Context, projectId, sha, ciToken string, statusData domain.GitlabCommitStatus) (any, error) {

	logging.Info("Creating commit status", "sha", sha)
	httpClient, err := inter.createHttpClient(ciToken)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(CommitStatusRequest{
		State:       statusData.Status,
		Name:        statusData.Name,
		TargetUrl:   statusData.TargetUrl,
		Description: statusData.Description,
		Coverage:    statusData.Coverage,
		PipelineId:  &statusData.PipelineId,
	})
	if err != nil {
		return nil, err
	}

	payload := string(data)

	var response any

	err0 := httpClient.Post(ctx, "/projects/"+(utils.EncodeURIComponent(projectId))+"/statuses/"+sha, url.Values{}, payload, &response)
	if err0 != nil {
		logging.Error("error creating commit status", err0.Error())
		return nil, err0
	}

	logging.Info("created commit status for", "sha", sha, "response", response)

	return response, nil
}

func (inter *Interactor) CreateOrUpdateMergeRequestComment(ctx context.Context, projectId, mergeRequestId, comment, ciToken string) (*domain.GitLabNote, error) {

	logging.Info("Creating comment with sonar report")

	httpClient, err := inter.createHttpClient(ciToken)

	if err != nil {
		return nil, err
	}

	var notesResponse []NoteResponse

	if err0 := httpClient.Get(
		ctx,
		"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId+"/notes",
		url.Values{},
		&notesResponse,
	); err0 != nil {
		logging.Error("échec de récupération des notes de la MR/PR", "mrIID", mergeRequestId)
		return nil, err0
	}
	var existingComment *NoteResponse
	if notesResponse != nil {
		for _, note := range notesResponse {
			if strings.Contains(note.Body, "SonarQube Analysis Report") && !note.System {
				existingComment = &note
				break
			}
		}
	}

	if existingComment != nil {
		logging.Info("Deleting de existing comment to pin up sonarqube analysis report")
		errdel := httpClient.Delete(
			ctx,
			"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId+"/notes/"+strconv.Itoa(existingComment.ID),
			nil,
			"",
			nil,
		)
		logging.Info("Deleted existing sticky comment")
		if errdel != nil {
			logging.Info("échec de suppression du dernier commentaire type Analyse Sonar (commit note)", "commentId", existingComment.ID)
			return nil, errdel
		}

		response, errc := createNewComment(ctx, httpClient, projectId, mergeRequestId, comment)

		if errc != nil {
			logging.Info("échec de création d'un nouveau commentaire Analyse Sonar (commit note)", "commentId", existingComment.ID)
			return nil, errc
		}
		logging.Info("spin up existing comment on MR !", mergeRequestId)
		return toDomain(*response), nil

	}

	resp, _ := createNewComment(ctx, httpClient, projectId, mergeRequestId, comment)
	return toDomain(*resp), nil
}

func (inter *Interactor) GetMergeRequest(ctx context.Context, projectId, mergeRequestId, ciToken string) (*domain.GitLabMergeRequest, error) {
	httpClient, err := inter.createHttpClient(ciToken)
	if err != nil {
		logging.Error("Error creating httpclient")
		return nil, err
	}

	var response MergeRequestResponse

	if err0 := httpClient.Get(
		ctx,
		"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId,
		nil,
		&response,
	); err0 != nil {
		logging.Error("Error getting MR","mrIID", mergeRequestId, err0)
		return nil, err0
	}

	return &domain.GitLabMergeRequest{
		ID:             response.ID,
		IID:            response.IID,
		ProjectID:      response.ProjectID,
		Title:          response.Title,
		Description:    response.Description,
		State:          response.State,
		CreatedAt:      response.CreatedAt,
		UpdatedAt:      response.UpdatedAt,
		MergedAt:       response.MergedAt,
		ClosedAt:       response.ClosedAt,
		TargetBranch:   response.TargetBranch,
		SourceBranch:   response.SourceBranch,
		SHA:            response.SHA,
		MergeCommitSHA: response.MergeCommitSHA,
		DiffRefs: domain.DiffRefs{
			BaseSHA:  response.DiffRefs.BaseSHA,
			HeadSHA:  response.DiffRefs.HeadSHA,
			StartSHA: response.DiffRefs.StartSHA,
		},
	}, nil
}

func createNewComment(ctx context.Context, httpClient *http.Client, projectId, mergeRequestId, comment string) (*NoteResponse, error) {
	data, _ := json.Marshal(map[string]string{"body": comment})
	var response NoteResponse
	err := httpClient.Post(
		ctx,
		"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId+"/notes",
		url.Values{},
		string(data),
		&response,
	)
	if err != nil {
		logging.Info("échec création du commentaire (commit note)", "mrIId", mergeRequestId, err)
		return nil, err
	}
	logging.Info("created new comment on MR !", "mrIID", mergeRequestId)
	return &response, nil
}

func toDomain(resp NoteResponse) *domain.GitLabNote {
	return &domain.GitLabNote{
		ID:   resp.ID,
		Body: resp.Body,
		Author: domain.GitLabAuthor{
			ID:       resp.Author.ID,
			Username: resp.Author.Username,
			Name:     resp.Author.Name,
		},
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
		System:    resp.System,
	}
}

var _ interactor.GitlabInteractor = (*Interactor)(nil)
