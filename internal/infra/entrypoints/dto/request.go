package dto

import (
	_ "github.com/go-playground/validator/v10"
)

type WebhookRequestDto struct {
	TaskID       string        `json:"taskId,omitempty"`
	TaskStatus   string        `json:"taskStatus,omitempty"`
	SonarProject SonarProject  `json:"sonarProject"`
	GitLab       Gitlab        `json:"gitlab"`
	MergeRequest *MergeRequest `json:"mergeRequest,omitempty"`
}

type SonarProject struct {
	Key string `json:"key"`
}

type Gitlab struct {
	ProjectID string  `json:"projectId"`
	CIToken   *string `json:"ciToken,omitempty"`
	Branch    string  `json:"branchName"`
	BranchUrl *string `json:"branchUrl,omitempty"`
	CommitSha string  `json:"commitSha"`
}

type MergeRequest struct {
	IID int `json:"iid"`
}

type VcsExtraData struct {
	Key   string `json:"key" validate:"required"`
	Value any    `json:"value" validate:"required"`
}
type VcsCreateCommentRequestDTO struct {
	ProjectId     string         `json:"project_id" validate:"required"`
	PullrequestId string         `json:"pullrequest_id" validate:"required"`
	CommentBody   string         `json:"comment_body" validate:"required,lte=1000000"`
	Provider      string         `json:"provider" validate:"required,oneof=gitlab github bb gitea"`
	Extra         []VcsExtraData `json:"extra,omitempty"`
}
