package sonar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/infra/http"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/logging"
	"strconv"
	"strings"
)

type Interactor struct {
	interactor.SonarInteractor
	config configs.Config
}

func New(cfg configs.Config) *Interactor {
	return &Interactor{
		SonarInteractor: (*Interactor)(nil),
		config:          cfg,
	}
}

var sonarClient *http.Client

var _ interactor.SonarInteractor = (*Interactor)(nil)

func (inter *Interactor) createHttpClient() (*http.Client, error) {
	if sonarClient != nil {
		return sonarClient, nil
	}
	baseUrl := inter.config.SonarBaseUrl
	token := inter.config.SonarToken
	if baseUrl == "" || token == "" {
		return nil, errors.New("variable d'environnement SONARQUBE_URL ou SONARQUBE_TOKEN manquante")
	}
	return http.NewClientHttp(strings.TrimRight(baseUrl, "/"), token, "sonar", inter.config.SonarCACert), nil
}

func (inter *Interactor) GetTaskDetails(ctx context.Context, taskId string) (*domain.SonarTaskDetails, error) {

	sonarClient, err := inter.createHttpClient()
	if err != nil {
		return nil, err
	}
	var response TaskResponse

	if err0 := sonarClient.Get(ctx, "/ce/task", url.Values{"id": {taskId}}, &response); err0 != nil {
		logging.Error("échec de récupération de la dernière analyse sonar", "taskId", taskId, "error", err0)
		return nil, fmt.Errorf("échec de récupération de la dernière analyse sonar [ID=%s], [Status=%d]", taskId, err0.StatusCode)
	}

	return &domain.SonarTaskDetails{
		Id:         taskId,
		AnalysisId: response.Task.AnalysisID,
		Status:     domain.TaskStatus(response.Task.Status),
	}, nil
}

func (inter *Interactor) GetAnalysisDetails(ctx context.Context, projectKey, branch, taskId string, taskStatus domain.TaskStatus) (*domain.AnalysisDetails, error) {
	var analysisId string
	var actualTaskId = taskId
	_status := taskStatus

	if len(actualTaskId) == 0 {
		latestAnalysis, err := inter.GetLatestAnalysis(ctx, projectKey, branch)
		if err != nil {
			return nil, err
		}
		analysisId = latestAnalysis.AnalysisId
		actualTaskId = latestAnalysis.TaskId
		_status = latestAnalysis.TaskStatus
	} else {
		task, err := inter.GetTaskDetails(ctx, actualTaskId)
		if err != nil {
			return nil, err
		}
		analysisId = task.AnalysisId
		_status = task.Status
	}

	sonarClient, err := inter.createHttpClient()
	if err != nil {
		panic(err)
	}

	// Get Quality gate status
	var qgResponse QualityGateStatusResponse

	if err0 := sonarClient.Get(ctx,
		"/qualitygates/project_status",
		url.Values{"analysisId": {analysisId}}, &qgResponse); err0 != nil {
		logging.Error("échec récupération project_status", "analysisId", analysisId, "erreur", err0)
		return nil, fmt.Errorf("impossible de récupérer l'analyse [ID=%s], le server sonarqube a répondu %d", analysisId, err0.StatusCode)
	}

	if qgResponse.ProjectStatus.Status == QG_NONE {
		return nil, fmt.Errorf("no quality gate associated with the analysis %s", analysisId)
	}

	// Get metrics
	var metricResponse MeasuresResponse
	metricKeys := strings.Join([]string{
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
	}, ",")

	if err1 := sonarClient.Get(ctx, "/measures/component",
		url.Values{"component": {projectKey}, "metricKeys": {metricKeys}}, &metricResponse); err1 != nil {
		logging.Error("échec de récupération des mesures (métriques)", "taskId", taskId, "error", err1)
		return nil, fmt.Errorf("échec de récupération des mesures (métriques)")
	}

	issuesResponse, err := inter.getAllBranchIssues(ctx, projectKey, branch)
	if err != nil {
		return nil, err
	}

	logging.Info("quality gate récupéré",
		"status", qgResponse.ProjectStatus.Status,
		"nbConditions", len(qgResponse.ProjectStatus.Conditions),
	)

	metrics := map[string]any{}

	if metricResponse.Component.Measures != nil {
		for _, measure := range metricResponse.Component.Measures {
			if measure.Period != nil {
				metrics[measure.MetricName] = measure.Period.Value
			} else {
				metrics[measure.MetricName] = measure.Value
			}
		}
	}

	dashboardUrl := strings.TrimRight(inter.config.SonarBaseUrl, "/api") + "/dashboard?codeScope=overall&id=" + projectKey + "&branch=" + branch

	var newIssues []any

	for _, issue := range issuesResponse.Issues {
		if issue.IsNew {
			newIssues = append(newIssues, issue)
		}
	}

	conds := make([]domain.QualityGateCondition, len(qgResponse.ProjectStatus.Conditions))
	for _, cond := range qgResponse.ProjectStatus.Conditions {
		conds = append(conds, domain.QualityGateCondition{
			Status:         string(cond.Status),
			MetricKey:      cond.MetricKey,
			Comparator:     cond.Comparator,
			PeriodIndex:    nil,
			ErrorThreshold: cond.ErrorThreshold,
			ActualValue:    cond.ActualValue,
		})
	}

	issues := make([]sonar.Issue, len(issuesResponse.Issues))

	for _, iss := range issuesResponse.Issues {
		issues = append(issues, sonar.Issue{
			Key:       iss.Key,
			Rule:      iss.Rule,
			Severity:  iss.Severity,
			Component: iss.Component,
			Project:   iss.Project,
			Line:      iss.Line,
			Message:   iss.Message,
			Status:    iss.Status,
			Type:      iss.Type,
			IsNew:     iss.IsNew,
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
			Status:     string(qgResponse.ProjectStatus.Status),
			Conditions: conds,
		},
		TotalIssues: issuesResponse.Total,
		TaskId:      actualTaskId,
		NewIssues:   len(newIssues),
		TaskStatus:  _status,
		Issues:      issues,
	}, nil
}

func (inter *Interactor) GetLatestAnalysis(ctx context.Context, projectKey, branch string) (*domain.AnalysisDetails, error) {

	sonarClient, e := inter.createHttpClient()
	if e != nil {
		panic(httpx.ErrInternal)
	}

	var response TasksResponse
	if er := sonarClient.Get(ctx, "/ce/activity", url.Values{
		"component": {projectKey},
		"branch":    {branch},
		"ps":        {"1"},
	}, &response); er != nil {
		logging.Error("échec de récupération de la dernière activité Sonar", "branch", branch, "error", er)
		return nil, fmt.Errorf("impossible de récupérer la dernière activité Sonar: [%d]", er.StatusCode)
	}

	lastTask := response.Tasks[0]
	if len(response.Tasks) <= 0 {
		return nil, errors.New("no analysis found for the project key (ComponentKey): " + projectKey)
	}
	analysis, err := inter.GetAnalysisDetails(ctx, projectKey, branch, lastTask.AnalysisID, domain.TaskStatus(lastTask.Status))

	if err != nil {
		logging.Error("échec de récupération de la dernière analyse sonar", "analysisId", lastTask.AnalysisID, "error", err)
		return nil, err
	}

	return analysis, nil
}

// go:deprecated
func (inter *Interactor) GetLastTask(ctx context.Context, projectKey, branch string) (*domain.SonarTaskDetails, error) {
	sonarClient, e := inter.createHttpClient()
	if e != nil {
		return nil, e
	}

	var response TasksResponse
	if er := sonarClient.Get(ctx, "/ce/activity", url.Values{
		"component": {projectKey},
		"branch":    {branch},
		"ps":        {"1"},
	}, &response); er != nil {
		logging.Error("échec de récupération de la dernière activité Sonar", "branch", branch, "error", er)
		return nil, fmt.Errorf("échec de récupération de la dernière activité Sonar: [%d]", er.StatusCode)
	}

	return nil, nil
}

func (inter *Interactor) GetMeasures(ctx context.Context, projectId, branch string) (*sonar.Measures, error) {
	return nil, nil
}

func (inter *Interactor) GetIssues(ctx context.Context, projectId, branch string) ([]sonar.Issue, error) {

	return nil, nil
}

func (inter *Interactor) getIssuesIteration(client *http.Client, ctx context.Context, projectKey, branch string, pageNumber, pageSize int) (*IssuesResponse, error) {
	var issuesResponse IssuesResponse
	ps := utils.Ternary(pageSize == 0, 100, pageSize)
	p := utils.Ternary(pageNumber == 0, 1, pageNumber)

	issueParams := url.Values{
		"componentKeys": {projectKey},
		"resolved":      {"false"},
		"ps":            {strconv.Itoa(ps)},
		"p":             {strconv.Itoa(p)},
		"facets":        {"severities,types,rules,tags"},
		"statuses":      {"OPEN,CONFIRMED,REOPENED"},
		"severities":    {"BLOCKER,CRITICAL,MAJOR,MINOR"},
	}

	// paramètre Developer Edition+, ignioré silencieusement sans erreur
	if inter.config.SonarEdition != configs.SONARQUBE_CE {
		issueParams.Add("branch", branch)
	}

	if err2 := client.Get(ctx, "/issues/search", issueParams, &issuesResponse); err2 != nil {
		return nil, fmt.Errorf("impossible de récupérer les issues sonar: [%d]", err2.StatusCode)
	}
	return &issuesResponse, nil
}

func (inter *Interactor) getAllBranchIssues(ctx context.Context, projectKey, branch string) (*IssuesResponse, error) {
	sonarClient, e := inter.createHttpClient()
	if e != nil {
		panic(e)
	}
	const pageSize = 50
	var mergeComponents = func(existingComponents, incomingComponents []IssueSearchComponent) []IssueSearchComponent {
		seen := make(map[string]struct{}, len(existingComponents))
		for _, c := range existingComponents {
			seen[c.Key] = struct{}{}
		}

		for _, c := range incomingComponents {
			if _, ok := seen[c.Key]; !ok {
				existingComponents = append(existingComponents, c)
				seen[c.Key] = struct{}{}
			}
		}

		return existingComponents
	}

	first, err := inter.getIssuesIteration(sonarClient, ctx, projectKey, branch, 1, pageSize)
	if err != nil {
		return nil, err
	}

	result := &IssuesResponse{
		Total:       first.Total,
		P:           first.P,
		Ps:          first.Ps,
		EffortTotal: first.EffortTotal,
		Issues:      first.Issues,
		Components:  first.Components,
		Facets:      first.Facets, // facets calculées sur le total, pas juste la page
	}

	totalFetched := len(first.Issues)
	page := 2

	for totalFetched < first.Total {
		if (page-1)*pageSize >= 10_000 {
			return nil, fmt.Errorf("issues: p×ps limit reached (10 000), filter issues to paginate further")
		}

		r, er := inter.getIssuesIteration(sonarClient, ctx, projectKey, branch, page, pageSize)
		if er != nil {
			return nil, er
		}
		result.Issues = append(result.Issues, r.Issues...)
		result.Components = mergeComponents(result.Components, r.Components)
		totalFetched += len(r.Issues)
		page++
	}

	return result, nil
}

type sonarErrorResponse struct {
	Errors []struct {
		Msg string `json:"msg"`
	} `json:"errors"`
}

func extractSonarError(body []byte) string {
	body = []byte(strings.TrimSpace(string(body)))

	if len(body) == 0 {
		return ""
	}

	// SonarQube error response
	var response sonarErrorResponse

	if err := json.Unmarshal(body, &response); err == nil {
		if len(response.Errors) > 0 {
			messages := make([]string, 0, len(response.Errors))

			for _, err := range response.Errors {
				if msg := strings.TrimSpace(err.Msg); msg != "" {
					messages = append(messages, msg)
				}
			}

			if len(messages) > 0 {
				return strings.Join(messages, "; ")
			}
		}
	}

	// Fallback : réponse texte / HTML / proxy / gateway
	return string(body)
}
