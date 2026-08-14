package render

// Format identify a output format ("markdown", "html", "json", ...)
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
	FormatJSON     Format = "json"
)

func (f Format) String() string {
	return string(f)
}

// IsZero Indicate that the format is not specified
func (f Format) IsZero() bool {
	return f == ""
}
