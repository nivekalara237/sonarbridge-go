package dto

type WebhookRequestDto struct {
	TaskID       string        `json:"taskId"`
	SonarProject SonarProject  `json:"sonarProject"`
	GitLab       Gitlab        `json:"gitlab"`
	MergeRequest *MergeRequest `json:"mergeRequest"`
}

type SonarProject struct {
	Key string `json:"key"`
}

type Gitlab struct {
	ProjectID string  `json:"projectId"`
	CIToken   *string `json:"ciToken"`
	Branch    string  `json:"branchName"`
	BranchUrl *string `json:"branchUrl"`
	CommitSha string  `json:"commitSha"`
}

type MergeRequest struct {
	IID int `json:"iid"`
}
