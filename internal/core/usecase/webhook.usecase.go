package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/renderer"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/render"
	"strconv"
)

type Service struct {
	interactor.SonarInteractor
	interactor.GitlabInteractor
	interactor.ReportInteractor
	repository.Repository
	renderer *renderer.SonarReportRender
}

func (svc *Service) Execute(ctx context.Context, webhookData domain.SonarQubeWebhookPayload) (*domain.WebhookResponse, error) {

	var taskStatus domain.TaskStatus
	if webhookData.Status != "" {
		taskStatus = domain.TaskStatus(webhookData.Status)
	} else {
		taskStatus = domain.TASK_UNKNOWN
	}

	analysis, err := svc.SonarInteractor.GetAnalysisDetails(
		ctx,
		webhookData.SonarProject.Key,
		webhookData.Branch.Name,
		*webhookData.TaskID,
		taskStatus,
	)
	if err != nil {
		logging.Error("échec récupération analysisId", "taskId", webhookData.TaskID, err)
	}

	commitSha, _ := svc.getCommitSha(ctx, webhookData.GitLab.ProjectID, strconv.Itoa(webhookData.MergeRequest.IID), webhookData.Branch.Commit.SHA, webhookData.GitLab.CIToken)

	if commitSha == "" {
		logging.Error("échec de recuperation du commit sha pour la branch", webhookData.Branch.Name)
		return nil, errors.New("impossible de récupérer le commit SHA ou branc.commit.sha not fourni")
	}
	// mergeRequestId =

	if _, errcc := svc.GitlabInteractor.CreateCommitStatus(ctx, webhookData.GitLab.ProjectID, commitSha, "", domain.GitlabCommitStatus{
		Status:      utils.Ternary(analysis.QualityGate.Status == "OK", "success", "failed"),
		Name:        "SonarQube Quality Gate",
		TargetUrl:   analysis.DashboardUrl,
		Description: fmt.Sprintf("SonarQube Quality Report: %s", analysis.QualityGate.Status),
		Coverage:    utils.ToFloat32OrZero(analysis.Metrics["coverage"]),
		PipelineId:  "",
	}); errcc != nil {
		logging.Error("échec de création du status du commit[", commitSha, "]", errcc)
		return nil, errcc
	}

	report := &sonar.Report{
		Status: aggregateAnalysisStatus(&domain.SonarTaskDetails{Status: analysis.TaskStatus}), // todo: fetch the real value from original task.status
		Analysis: sonar.Analysis{
			Key:       analysis.AnalysisId,
			Project:   webhookData.SonarProject.Key,
			Branch:    webhookData.Branch.Name,
			CommitSHA: webhookData.Branch.Commit.SHA,
		},
		QualityGate: aggregateQualityGateStatus(analysis.QualityGate.Status),
		Measures:    sonar.Measures{},
		Issues:      aggregateIssues(analysis.Issues),
		ReportURL:   analysis.DashboardUrl,
	}

	mergeable := slices.Contains([]string{"OK", "WARN"}, analysis.QualityGate.Status)

	if er := svc.ReportRepository.SaveReport(ctx, report); er != nil {
		logging.Error("The report have not saved")
	}

	renderEngine, err := renderer.NewRenderEngine()
	if err != nil {
		logging.Error("unabled to run render engine")
		return nil, err
	}

	content, e := renderEngine.RenderString(ctx, render.Request{
		Ref:    "embed://templates/report-2.md.tmpl",
		Format: render.FormatMarkdown,
		Data:   report,
	})

	if e != nil {
		logging.Error("enable to compile report template", "error", e)
		return nil, fmt.Errorf("unabled to compile template/render")
	}

	if webhookData.MergeRequest != nil {
		if _, errmg := svc.GitlabInteractor.CreateOrUpdateMergeRequestComment(
			ctx,
			webhookData.GitLab.ProjectID,
			strconv.Itoa(webhookData.MergeRequest.IID),
			content,
			"",
		); errmg != nil {
			logging.Error("échec post commentaire GitLab", "mrIID", webhookData.MergeRequest.IID, "erreur", errmg)
			return nil, errors.New("erreur lors du post du commentaire GitLab")
		}
	} else {
	}

	return &domain.WebhookResponse{
		Mergeable:         mergeable,
		Received:          true,
		QualityGateStatus: analysis.QualityGate.Status,
	}, nil
}

func (svc *Service) getCommitSha(ctx context.Context, gitlabProjectId, mergeRequestId, branchCommitSha string, ciToken string) (string, error) {
	if branchCommitSha != "" {
		return branchCommitSha, nil
	}

	mr, err := svc.GitlabInteractor.GetMergeRequest(ctx, gitlabProjectId, mergeRequestId, ciToken)
	if err != nil {
		logging.Error("échec de recuperation de la MR GitLab", "mrIID", mergeRequestId, "erreur", err)
		return "", err
	}

	return func(mr domain.GitLabMergeRequest) string {
		if mr.SHA != "" {
			return mr.SHA
		}
		if mr.MergeCommitSHA != nil && *mr.MergeCommitSHA != "" {
			return *mr.MergeCommitSHA
		}
		if mr.DiffRefs.HeadSHA != "" {
			return mr.DiffRefs.HeadSHA
		}
		return ""
	}(*mr), nil
}

func aggregateIssues(issues []sonar.Issue) sonar.IssueSummary {
	if len(issues) == 0 {
		return sonar.IssueSummary{}
	}
	var summary sonar.IssueSummary

	for _, issue := range issues {

		switch issue.Severity {
		case sonar.SeverityBlocker:
			summary.Blocker++
		case sonar.SeverityCritical:
			summary.Critical++
		case sonar.SeverityMajor:
			summary.Major++
		case sonar.SeverityMinor:
			summary.Minor++
		case sonar.SeverityInfo:
			summary.Info++
		}

		if !issue.IsNew {
			continue
		}

		summary.NewIssues++

		switch issue.Type {
		case sonar.IssueBug:
			summary.NewBugs++
		case sonar.IssueCodeSmell:
			summary.NewCodeSmells++
		case sonar.IssueVulnerability:
			summary.NewVulnerabilities++
		}
	}

	// summary.NewIssues =

	return summary
}

func aggregateAnalysisStatus(task *domain.SonarTaskDetails) sonar.AnalysisStatus {
	switch task.Status {
	case "SUCCESS":
		return sonar.AnalysisSuccess
	case "FAILED":
		return sonar.AnalysisFailed
	default:
		return sonar.AnalysisStatus(task.Status)
	}
}

func aggregateQualityGateStatus(status string) sonar.QualityGateStatus {
	switch status {
	case "SUCCESS", "OK":
		return sonar.QualityGatePassed
	case "FAILED", "ERROR":
		return sonar.QualityGateFailed
	case "WARN":
		return sonar.QualityGateWARNED
	default:
		return sonar.QualityGateUnknown
	}
}

func buildReportURL(
	analysis *sonar.Analysis,
) string {

	if analysis.Project == "" {
		return ""
	}

	return fmt.Sprintf(
		"%s/dashboard?id=%s",
		analysis.Branch,
		analysis.Project,
	)
}
