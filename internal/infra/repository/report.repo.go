package repository

import (
	"fmt"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/interactor"
	"strings"
)

type Interactor struct {
	interactor.ReportInteractor
}

func (inter *Interactor) FormatMarkdown(details domain.AnalysisDetails) string {
	var sb strings.Builder
	icon := "✅"
	if details.QualityGate.Status != "OK" {
		icon = "❌"
	}

	fmt.Fprintf(&sb, "## %s Rapport SonarQube - Quality Gate: %s\n\n", icon, details.QualityGate.Status)
	sb.WriteString("| Métrique | Statut | Valeur réelle | Seuil |\n")
	sb.WriteString("|---|---|---|---|\n")

	for _, c := range details.QualityGate.Conditions {
		condIcon := "✅"
		if c.Status != "OK" {
			condIcon = "❌"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %s %s |\n",
			c.MetricKey, condIcon, c.ActualValue, c.Comparator, c.ErrorThreshold)
	}

	return sb.String()
}

func (inter *Interactor) IsMergeable(status domain.AnalysisDetails) bool {
	return status.QualityGate.Status == "OK"
}
