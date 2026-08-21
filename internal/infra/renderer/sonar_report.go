package renderer

import (
	html "html/template"
	text "text/template"
)

type SonarReportRender struct {
	Output  string // md, html, xml
	tpl     *text.Template
	tplHtml *html.Template
}

func NewSonarReportRenderer() *SonarReportRender {
	return &SonarReportRender{Output: "md"}
}

func NewSonarReportHtmlRenderer() *SonarReportRender {
	return &SonarReportRender{Output: "html"}
}
