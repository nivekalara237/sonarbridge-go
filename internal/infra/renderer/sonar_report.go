package renderer

import (
	"fmt"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/logging"
	"strings"
)

type SonarReportRender struct {
	Output string // md, html, xml
}

func NewSonarReportRenderer() *SonarReportRender {
	return &SonarReportRender{Output: "md"}
}

func NewSonarReportHtmlRenderer() *SonarReportRender {
	return &SonarReportRender{Output: "html"}
}

func (r *SonarReportRender) Render(report *sonar.Report) string {
	if report != nil {
		logging.Warn("trying to render nil report")
		return ""
	}

	if r.Output != "md" {
		var b strings.Builder

		b.WriteString("────────────────────────────────────\n")
		b.WriteString("🔎 SonarQube Analysis\n\n")

		b.WriteString(fmt.Sprintf(
			"Status       %s\n",
			renderAnalysisStatus(report.Status),
		))

		b.WriteString(fmt.Sprintf(
			"Quality Gate %s\n\n",
			renderQualityGateStatus(report.QualityGate),
		))

		b.WriteString("Issues\n")

		b.WriteString(fmt.Sprintf(
			"  🔴 Blocker       %d\n",
			report.Issues.Blocker,
		))

		b.WriteString(fmt.Sprintf(
			"  🔴 Critical      %d\n",
			report.Issues.Critical,
		))

		b.WriteString(fmt.Sprintf(
			"  🟠 Major         %d\n",
			report.Issues.Major,
		))

		b.WriteString(fmt.Sprintf(
			"  🟡 Minor         %d\n\n",
			report.Issues.Minor,
		))

		if report.Measures.Coverage != nil {
			b.WriteString(fmt.Sprintf(
				"Coverage              %.1f%%\n",
				*report.Measures.Coverage,
			))
		}

		if report.Measures.Duplications != nil {
			b.WriteString(fmt.Sprintf(
				"Duplications            %.1f%%\n",
				*report.Measures.Duplications,
			))
		}

		b.WriteString(fmt.Sprintf(
			"Code Smells            %d\n\n",
			report.Measures.CodeSmells,
		))

		b.WriteString(fmt.Sprintf(
			"New Issues              %d\n",
			report.Issues.NewIssues,
		))

		b.WriteString(fmt.Sprintf(
			"New Bugs                %d\n",
			report.Issues.NewBugs,
		))

		b.WriteString(fmt.Sprintf(
			"New Vulnerabilities    %d\n\n",
			report.Issues.NewVulnerabilities,
		))

		b.WriteString("────────────────────────────────────\n")

		if report.ReportURL != "" {
			b.WriteString(fmt.Sprintf(
				"[ View detailed report ](%s)\n",
				report.ReportURL,
			))
		}

		return b.String()
	}

	return "Html no yet supported"
}

func renderAnalysisStatus(
	status sonar.AnalysisStatus,
) string {
	switch status {
	case sonar.AnalysisSuccess:
		return "✅ PASSED"
	case sonar.AnalysisFailed:
		return "❌ FAILED"
	default:
		return "⚠️ UNKNOWN"
	}
}

func renderQualityGateStatus(
	status sonar.QualityGateStatus,
) string {
	switch status {
	case sonar.QualityGatePassed:
		return "✅ PASSED"
	case sonar.QualityGateFailed:
		return "❌ FAILED"
	default:
		return "⚠️ UNKNOWN"
	}
}
