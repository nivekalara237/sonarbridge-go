package markdown

import (
	"embed"
)

//go:embed templates/*
var templateFS embed.FS

type MarkdownRenderer struct{}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func (r *MarkdownRenderer) GetTmplFS() embed.FS {
	return templateFS
}
