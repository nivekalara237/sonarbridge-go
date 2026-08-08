package repository

import "sonarbridge-go/internal/infra/repository/report"

type Repository struct {
	ReportRepository *report.Repository
}

func New(
	report *report.Repository,
) *Repository {
	return &Repository{
		ReportRepository: report,
	}
}
