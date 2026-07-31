package sonar

type Severity string

const (
	SeverityBlocker  Severity = "BLOCKER"
	SeverityCritical Severity = "CRITICAL"
	SeverityMajor    Severity = "MAJOR"
	SeverityMinor    Severity = "MINOR"
	SeverityInfo     Severity = "INFO"
)

type IssueType string

const (
	IssueTypeBug           IssueType = "BUG"
	IssueTypeVulnerability IssueType = "VULNERABILITY"
	IssueTypeCodeSmell     IssueType = "CODE_SMELL"
)

type TaskResponse struct {
	Task struct {
		AnalysisID string `json:"analysisId"`
		Status     string `json:"status"`
	} `json:"task"`
}

type TasksResponse struct {
	Tasks []struct {
		AnalysisID string `json:"analysisId"`
		Status     string `json:"status"`
	} `json:"tasks"`
}

type QualityGateCondition struct {
	Status         string  `json:"status"`
	MetricKey      string  `json:"metricKey"`
	Comparator     string  `json:"comparator"`
	PeriodIndex    *string `json:"periodIndex,omitempty"`
	ErrorThreshold string  `json:"errorThreshold"`
	ActualValue    string  `json:"actualValue"`
}

type QualityGateStatusResponse struct {
	ProjectStatus struct {
		Status            string                 `json:"status"`
		Conditions        []QualityGateCondition `json:"conditions"`
		IgnoredConditions bool                   `json:"ignoredConditions"`
		Period            *struct {
			Index     int    `json:"index"`
			Mode      string `json:"mode"`
			Date      string `json:"date"`
			Parameter string `json:"parameter"`
		} `json:"period,omitempty"`
	} `json:"projectStatus"`
}

type MeasuresResponse struct {
	Component struct {
		Id        string `json:"id"`
		Key       string `json:"key"`
		Name      string `json:"name"`
		Qualifier string `json:"qualifier"`
		Measures  *[]struct {
			Metric    string `json:"metric"`
			Value     string `json:"value"`
			BestValue bool   `json:"bestValue"`
			Periods   []struct {
				Index     int    `json:"index"`
				Value     string `json:"value"`
				BestValue bool   `json:"bestValue"`
			} `json:"periods,omitempty"`
		} `json:"measures,omitempty"`
	} `json:"component"`
}

type TextRange struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartOffset int `json:"startOffset"`
	EndOffset   int `json:"endOffset"`
}

type IssuesItemResponse struct {
	Key               string     `json:"key"`
	Rule              string     `json:"rule"`
	Severity          Severity   `json:"severity"`
	Component         string     `json:"component"`
	Project           string     `json:"project"`
	Line              *int       `json:"line,omitempty"`
	Hash              *string    `json:"hash,omitempty"`
	TextRange         *TextRange `json:"textRange,omitempty"`
	Flows             []any      `json:"flows"`
	Status            string     `json:"status"`
	Message           string     `json:"message"`
	Effort            string     `json:"effort"`
	Debt              string     `json:"debt"`
	Assignee          *string    `json:"assignee,omitempty"`
	Author            *string    `json:"author,omitempty"`
	Tags              []string   `json:"tags"`
	CreationDate      string     `json:"creationDate"`
	UpdateDate        string     `json:"updateDate"`
	Type              IssueType  `json:"type"`
	Scope             string     `json:"scope"`
	QuickFixAvailable bool       `json:"quickFixAvailable"`
	IsNew             bool       `json:"isNew"`
}

type IssuesResponse struct {
	Total       int                  `json:"total"`
	P           float32              `json:"p"`
	Ps          float32              `json:"ps"`
	EffortTotal float32              `json:"effortTotal"`
	Issues      []IssuesItemResponse `json:"issues"`
	Components  []struct {
		Key       string `json:"key"`
		Enabled   bool   `json:"enabled"`
		Qualifier string `json:"qualifier"`
		Name      string `json:"name"`
		LongName  string `json:"longName"`
		Path      string `json:"path"`
	} `json:"components"`
}

type MetricValue interface {
	~string | ~int | ~int64 | ~float32 | ~float64
}
type Metrics[T MetricValue] map[string]T
