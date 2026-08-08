package sonar

type Measures struct {
	Coverage           *float64
	Duplications       *float64
	CodeSmells         int
	Bugs               int
	Vulnerabilities    int
	Violations         int
	NewIssues          int
	NewBugs            int
	NewVulnerabilities int
}

type MeasureItem struct {
	Name  string
	Value any
}
