package sonar

type Analysis struct {
	Key       string
	Project   string
	Branch    string
	CommitSHA string
	Date      string
}
