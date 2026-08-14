package dto

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
