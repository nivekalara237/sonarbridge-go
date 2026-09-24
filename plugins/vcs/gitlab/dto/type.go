package dto

type Note struct {
	ID              int64          `json:"id"`
	Type            string         `json:"type"`
	Body            string         `json:"body"`
	Author          *GitLabUser    `json:"author,omitempty"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	System          bool           `json:"system"`
	NoteableID      int64          `json:"noteable_id"`
	NoteableType    string         `json:"noteable_type"`
	ProjectID       int64          `json:"project_id"`
	CommitID        string         `json:"commit_id"`
	Position        map[string]any `json:"position"`
	Resolvable      bool           `json:"resolvable"`
	Resolved        bool           `json:"resolved"`
	ResolvedBy      *GitLabUser    `json:"resolved_by,omitempty"`
	ResolvedAt      *string        `json:"resolved_at,omitempty"`
	Suggestions     *Suggestion    `json:"suggestions,omitempty"`
	Confidential    bool           `json:"confidential"`
	Internal        bool           `json:"internal"`
	Imported        bool           `json:"imported"`
	ImportedFrom    string         `json:"imported_from"`
	NoteableIID     int64          `json:"noteable_iid"`
	CommandsChanges map[string]any `json:"commands_changes"`
}

type GitLabUser struct {
	ID               int64             `json:"id"`
	Username         string            `json:"username"`
	PublicEmail      string            `json:"public_email"`
	Name             string            `json:"name"`
	State            string            `json:"state"`
	Locked           bool              `json:"locked"`
	AvatarURL        string            `json:"avatar_url"`
	AvatarPath       string            `json:"avatar_path"`
	CustomAttributes []CustomAttribute `json:"custom_attributes"`
	WebURL           string            `json:"web_url"`
}

type CustomAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Suggestion struct {
	ID          int64  `json:"id"`
	FromLine    int    `json:"from_line"`
	ToLine      int    `json:"to_line"`
	Appliable   bool   `json:"appliable"`
	Applied     bool   `json:"applied"`
	FromContent string `json:"from_content"`
	ToContent   string `json:"to_content"`
}
