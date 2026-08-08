package domain

type SonarQubeWebhookPayload struct {
	TaskID       *string
	Status       WebhookStatus
	SonarProject Project
	GitLab       GitLab
	Branch       *Branch
	MergeRequest *MergeRequest
	Properties   map[string]string
}

type Project struct {
	Key  string
	Name string
}

type GitLab struct {
	ProjectID string
	CIToken   string
}

type Branch struct {
	Name   string
	IsMain bool
	URL    *string
	Commit *Commit
}

type Commit struct {
	SHA     string
	Message string
}

type MergeRequest struct {
	IID int
}

type WebhookResponse struct {
	Mergeable         bool
	Received          bool
	QualityGateStatus string
}

type WebhookStatus string

const (
	WebhookStatusSuccess WebhookStatus = "SUCCESS"
	WebhookStatusFailed  WebhookStatus = "FAILED"
)
