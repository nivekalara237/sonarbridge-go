package sonar

type IssueSeverity string

const (
	SeverityBlocker  IssueSeverity = "BLOCKER"
	SeverityCritical IssueSeverity = "CRITICAL"
	SeverityMajor    IssueSeverity = "MAJOR"
	SeverityMinor    IssueSeverity = "MINOR"
	SeverityInfo     IssueSeverity = "INFO"
)

type IssueType string

const (
	IssueBug           IssueType = "BUG"
	IssueVulnerability IssueType = "VULNERABILITY"
	IssueCodeSmell     IssueType = "CODE_SMELL"
	IssueCoverage      IssueType = "COVERAGE"
)

type Issue struct {
	Key       string
	Rule      string
	Severity  IssueType
	Component string
	Project   string
	Line      *int
	Message   string
	Status    string
	Type      IssueType
	IsNew     bool
}
