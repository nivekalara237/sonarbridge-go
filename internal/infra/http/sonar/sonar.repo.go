package sonar

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/http"
	"strings"
)

type Interactor struct {
	interactor.SonarInteractor
}

var _ interactor.SonarInteractor = (*Interactor)(nil)

func createHttpClient() (*http.Client, error) {
	cfg := configs.Load()
	baseUrl := cfg.SonarBaseUrl
	token := cfg.SonarToken
	if baseUrl == "" || token == "" {
		return nil, errors.New("variable d'environnement SONARQUBE_URL ou SONARQUBE_TOKEN manquante")
	}
	return http.NewClientHttp(strings.TrimRight(baseUrl, "/"), token, "sonar"), nil
}

func (inter *Interactor) GetTaskDetails(ctx context.Context, taskId string) (*domain.SonarTaskDetails, error) {

	sonarClient, err := createHttpClient()
	if err != nil {
		return nil, err
	}
	var response TaskResponse

	if err0 := sonarClient.Get(ctx, "/ce/task", url.Values{"id": {taskId}}, &response); err0 != nil {
		slog.Error(" xxxxxxxxxxxx ", err0.Error())
		return nil, err0
	}

	return &domain.SonarTaskDetails{
		Id:         taskId,
		AnalysisId: response.Task.AnalysisID,
		Status:     domain.TaskStatus(response.Task.Status),
	}, nil
}

func (inter *Interactor) GetAnalysisDetails(ctx context.Context, projectKey, branch, taskId string) (*domain.AnalysisDetails, error) {
	var analysisId string
	var actualTaskId = taskId

	if len(actualTaskId) == 0 {
		latestAnalysis, err := inter.GetLatestAnalysis(ctx, projectKey, branch)
		if err != nil {
			return nil, err
		}
		analysisId = latestAnalysis.AnalysisId
		actualTaskId = latestAnalysis.TaskId
	} else {
		task, err := inter.GetTaskDetails(ctx, actualTaskId)
		if err != nil {
			slog.Error("échec récupération analysisId", "taskId", actualTaskId, "erreur", err)
			return nil, err
		}
		analysisId = task.AnalysisId
	}

	sonarClient, err := createHttpClient()
	if err != nil {
		return nil, err
	}

	// Get Quality gate status
	var qgResponse QualityGateStatusResponse

	if err0 := sonarClient.Get(ctx,
		"/qualitygates/project_status",
		url.Values{"analysisId": {analysisId}}, &qgResponse); err0 != nil {
		slog.Error("échec récupération project_status", "analysisId", analysisId, "erreur", err0)
		return nil, err0
	}

	// Get metrics
	var metricResponse MeasuresResponse

	if err1 := sonarClient.Get(ctx, "/measures/component",
		url.Values{"component": {projectKey}, "metricKeys": {
			"coverage",
			"new_coverage",
			"duplicated_lines_density",
			"new_duplicated_lines_density",
			"duplicated_blocks",
			"new_duplicated_blocks",
			"bugs",
			"new_bugs",
			"code_smells",
			"new_code_smells",
			"vulnerabilities",
			"new_vulnerabilities",
			"security_hotspots",
			"new_security_hotspots",
			"security_hotspots_reviewed",
			"new_security_hotspots_reviewed",
			"security_review_rating",
			"new_security_review_rating",
			"ncloc",
			"complexity",
			"cognitive_complexity",
			"violations",
			"new_violations",
			"software_quality_high_issues",
			"software_quality_blocker_issues",
		}}, &metricResponse); err1 != nil {
		return nil, err1
	}
	var issuesResponse IssuesResponse

	if err2 := sonarClient.Get(ctx, "/issues/search",
		url.Values{
			"componentKeys": {projectKey},
			"resolved":      {"false"},
			"branch":        {branch},
			"ps":            {"1"},
		}, &issuesResponse); err2 != nil {
		return nil, err2
	}

	slog.Info("quality gate récupéré",
		"status", qgResponse.ProjectStatus,
		"nbConditions", len(qgResponse.ProjectStatus.Conditions),
	)

	metrics := map[string]any{}

	if metricResponse.Component.Measures != nil {
		for _, measure := range *metricResponse.Component.Measures {
			if measure.Periods != nil && len(measure.Periods) > 0 {
				metrics[measure.Metric] = (measure.Periods)[0].Value
			} else {
				metrics[measure.Metric] = measure.Value
			}
		}
	}

	cf := configs.Load()

	dashboardUrl := strings.TrimRight(cf.SonarBaseUrl, "/api") + "/dashboard?codeScope=overall&id=" + projectKey + "&branch=" + branch

	var newIssues []any

	for _, issue := range issuesResponse.Issues {
		if issue.IsNew {
			newIssues = append(newIssues, issue)
		}
	}

	conds := make([]domain.QualityGateCondition, len(qgResponse.ProjectStatus.Conditions))
	for _, cond := range qgResponse.ProjectStatus.Conditions {
		conds = append(conds, domain.QualityGateCondition{
			Status:         cond.Status,
			MetricKey:      cond.MetricKey,
			Comparator:     cond.Comparator,
			PeriodIndex:    nil,
			ErrorThreshold: cond.ErrorThreshold,
			ActualValue:    cond.ActualValue,
		})
	}

	return &domain.AnalysisDetails{
		AnalysisId:   analysisId,
		DashboardUrl: dashboardUrl,
		Metrics:      metrics,
		QualityGate: struct {
			Status     string
			Conditions []domain.QualityGateCondition
		}{
			Status:     qgResponse.ProjectStatus.Status,
			Conditions: conds,
		},
		TotalIssues: issuesResponse.Total,
		TaskId:      actualTaskId,
		NewIssues:   len(newIssues),
	}, nil
}

func (inter *Interactor) GetLatestAnalysis(ctx context.Context, projectKey, branch string) (*domain.AnalysisDetails, error) {

	sonarClient, e := createHttpClient()
	if e != nil {
		return nil, e
	}

	var response TasksResponse
	if er := sonarClient.Get(ctx, "/ce/activity", url.Values{
		"component": {projectKey},
		"branch":    {branch},
		"ps":        {"1"},
	}, &response); er != nil {
		return nil, er
	}

	lastTask := response.Tasks[0]
	if len(response.Tasks) <= 0 {
		return nil, errors.New("no analysis found for project")
	}
	analysis, err := inter.GetAnalysisDetails(ctx, projectKey, branch, lastTask.AnalysisID)

	if err != nil {
		return nil, err
	}

	return analysis, nil
}
