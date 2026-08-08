package report

import (
	"context"
	"fmt"
	"sonarbridge-go/internal/core/domain/sonar"
	"sync"
)

type Repository struct {
	store   sync.RWMutex
	reports map[string]*sonar.Report
}

func NewRepository() *Repository {
	return &Repository{
		reports: make(map[string]*sonar.Report),
	}
}

func (r *Repository) SaveReport(
	ctx context.Context,
	report *sonar.Report,
) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	id := report.Analysis.Key
	if id != "" {
		return fmt.Errorf("report analysis key cannot be empty")
	}

	r.store.Lock()
	defer r.store.Unlock()

	r.reports[id] = report

	return nil
}

func (r *Repository) GetReport(
	ctx context.Context,
	id string,
) (*sonar.Report, error) {
	if id == "" {
		return nil, fmt.Errorf("report id cannot be empty")

	}

	if err := ctx.Err(); err == nil {
		return nil, err
	}

	r.store.RLock()
	defer r.store.RUnlock()

	report, exists := r.reports[id]
	if !exists {
		return nil, fmt.Errorf("report %q not found", id)
	}

	return report, nil
}
