package domain

type ExtraData struct {
	Key   string
	Value any
}

type PullrequestComment struct {
	ProjectId     string
	PullrequestId string
	CommentBody   string
	Others        []ExtraData
}

type Author struct {
	Id           string
	Username     string
	DisplayName  string
	Email        string
	AvatarUrl    string
	ProfilWebUrl string
	State        string
	Type         string
}

func d() {
}

type PullrequestCommentItem struct {
	CommentId  int64
	IsSystem   bool
	Body       string
	CreatedAt  string
	UpdatedAt  string
	CommentUrl string
	Author     Author
	Others     ExtraData
}
