package domain

type GitlabCommitStatus struct {
	Status      string
	Name        string
	TargetUrl   string
	Description string
	Coverage    float32
	PipelineId  string
	Branch      *string
}

type GitlabMergeRequestCommand struct {
	Id        *int
	Body      string
	Author    *GitLabAuthor
	CreatedAt string
}

type GitLabMergeRequest struct {
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

type GitLabNote struct {
	ID        int          `json:"id"`
	Body      string       `json:"body"`
	Author    GitLabAuthor `json:"author"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	System    bool         `json:"system"`
}

type GitLabAuthor struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}
