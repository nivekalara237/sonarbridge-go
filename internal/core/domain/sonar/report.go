package sonar

type QualityGateStatus string

const (
	QualityGatePassed  QualityGateStatus = "PASSED"
	QualityGateFailed  QualityGateStatus = "FAILED"
	QualityGateWARNED  QualityGateStatus = "WARNING"
	QualityGateUnknown QualityGateStatus = "UNKNOWN"
)

type AnalysisStatus string

const (
	AnalysisSuccess AnalysisStatus = "SUCCESS"
	AnalysisFailed  AnalysisStatus = "FAILED"
)

type Report struct {
	Status      AnalysisStatus
	QualityGate QualityGateStatus
	Issues      IssueSummary
	Measures    Measures
	// Measures  []MeasureItem
	Analysis  Analysis
	ReportURL string
}

type QualityGateCondition struct {
	Metric    string
	Status    string
	Value     string
	Threshold string
	IsError   bool
}

type QualityGate struct {
	Status     QualityGateStatus
	Conditions []QualityGateCondition
}

type IssueSummary struct {
	Blocker            int
	Critical           int
	Major              int
	Minor              int
	Info               int
	NewIssues          int
	NewCodeSmells      int
	NewBugs            int
	NewVulnerabilities int
}
