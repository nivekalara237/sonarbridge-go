package gitlab

type CommitStatusRequest struct {
	State       string  `json:"state"`
	ProjectId   string  `json:"id"`
	Name        string  `json:"name"`
	BranchRef   *string `json:"ref"`
	TargetUrl   string  `json:"target_url"`
	Description string  `json:"description"`
	Coverage    float32 `json:"coverage"`
	PipelineId  *string `json:"pipeline_id"`
}

type NoteResponse struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Author    Author `json:"author"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	System    bool   `json:"system"`
}

type Author struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type MergeRequestResponse struct {
	ID             int      `json:"id"`
	IID            int      `json:"iid"`
	ProjectID      int      `json:"project_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	State          string   `json:"state"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	MergedAt       *string  `json:"merged_at"`
	ClosedAt       *string  `json:"closed_at"`
	TargetBranch   string   `json:"target_branch"`
	SourceBranch   string   `json:"source_branch"`
	SHA            string   `json:"sha"`
	MergeCommitSHA *string  `json:"merge_commit_sha"`
	DiffRefs       DiffRefs `json:"diff_refs"`
}

type DiffRefs struct {
	BaseSHA  string `json:"base_sha"`
	HeadSHA  string `json:"head_sha"`
	StartSHA string `json:"start_sha"`
}
