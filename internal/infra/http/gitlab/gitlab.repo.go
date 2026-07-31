package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/http"
	"sonarbridge-go/internal/infra/utils"
	"strconv"
	"strings"
)

type Interactor struct {
	interactor.GitlabInteractor
}

func createHttpClient(ciToken string) (*http.Client, error) {
	cfg := configs.Load()
	baseUrl := cfg.GitlabBaseUrl
	token := cfg.GitlabToken
	if baseUrl == "" || token == "" {
		return nil, errors.New("variable d'environnement GITLAB_API_URL ou GITLAB_TOKEN manquante")
	}
	if strings.TrimSpace(ciToken) != "" {
		token = "ci;" + token
	}
	return http.NewClientHttp(strings.TrimRight(baseUrl, "/"), token, "gitlab"), nil
}

func (inter *Interactor) CreateCommitStatus(ctx context.Context, projectId, sha, ciToken string, statusData domain.GitlabCommitStatus) (any, error) {

	httpClient, err := createHttpClient(ciToken)
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
		log.Fatalf("error creating commit status: %s", err0.Error())
		return nil, err0
	}

	log.Println("created commit status for", sha, response)

	return response, nil
}

func (inter *Interactor) CreateOrUpdateMergeRequestComment(ctx context.Context, projectId, mergeRequestId, comment, ciToken string) (*domain.GitLabNote, error) {

	httpClient, err := createHttpClient(ciToken)

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
		log.Println("Deleting de existing comment to pin up sonarqube analysis report")
		errdel := httpClient.Delete(
			ctx,
			"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId+"/notes/"+strconv.Itoa(existingComment.ID),
			nil,
			"",
			nil,
		)
		log.Println("Deleted existing sticky comment")
		if errdel != nil {
			return nil, errdel
		}

		response, errc := createNewComment(ctx, httpClient, projectId, mergeRequestId, comment)

		if errc != nil {
			return nil, errc
		}
		log.Println("spin up existing comment on MR !", mergeRequestId)
		return toDomain(*response), nil

	}

	resp, xdd := createNewComment(ctx, httpClient, projectId, mergeRequestId, comment)
	fmt.Println("**********************************")
	fmt.Println(resp)
	fmt.Println(xdd)
	fmt.Println("**********************************")
	return toDomain(*resp), nil
}

func (inter *Interactor) GetMergeRequest(ctx context.Context, projectId, mergeRequestId, ciToken string) (*domain.GitLabMergeRequest, error) {
	httpClient, err := createHttpClient(ciToken)
	if err != nil {
		log.Println("Error creating httpclient")
		return nil, err
	}

	var response MergeRequestResponse

	if err0 := httpClient.Get(
		ctx,
		"/projects/"+(utils.EncodeURIComponent(projectId))+"/merge_requests/"+mergeRequestId,
		nil,
		&response,
	); err0 != nil {
		log.Println("Error getting MR:", mergeRequestId)
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
		return nil, err
	}
	log.Println("created new comment on MR !", mergeRequestId)
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
