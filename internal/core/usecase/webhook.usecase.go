package usecase

import (
	"context"
	"errors"
	"fmt"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/domain/sonar"
	"sonarbridge-go/internal/core/interactor"
	"sonarbridge-go/internal/infra/renderer"
	"sonarbridge-go/internal/infra/repository"
	"sonarbridge-go/internal/infra/utils"
	"sonarbridge-go/internal/logging"
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

	logging.Info("Processing SonarQube webhook", map[string]any{
		"project": webhookData.SonarProject.Key,
		"branch":  webhookData.Branch.Name,
		"status":  webhookData.Status,
		"mr":      webhookData.MergeRequest,
	})

	analysis, err := svc.SonarInteractor.GetAnalysisDetails(
		ctx,
		webhookData.SonarProject.Key,
		webhookData.Branch.Name,
		*webhookData.TaskID,
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
		Description: fmt.Sprintf("SonarQube Quality gate: %s", analysis.QualityGate.Status),
		Coverage:    utils.ToFloat32OrZero(analysis.Metrics["coverage"]),
		PipelineId:  "",
	}); errcc != nil {
		logging.Error("échec de création du status du commit[", commitSha, "]", errcc)
		return nil, errcc
	}

	report := &sonar.Report{
		Status: sonar.AnalysisStatus(analysis.TaskId), // todo: fetch the real value from original task.status
		Analysis: sonar.Analysis{
			Key:       analysis.AnalysisId,
			Project:   webhookData.SonarProject.Key,
			Branch:    webhookData.Branch.Name,
			CommitSHA: webhookData.Branch.Commit.SHA,
		},
	}

	markdownReport := svc.ReportInteractor.FormatMarkdown(*analysis)
	mergeable := analysis.QualityGate.Status == "OK"

	if er := svc.ReportRepository.SaveReport(ctx, report); er != nil {
		logging.Error("The report have not saved")
	}

	reportMD := renderer.NewSonarReportRenderer().Render(report)

	fmt.Println(reportMD)

	if webhookData.MergeRequest != nil {
		if _, errmg := svc.GitlabInteractor.CreateOrUpdateMergeRequestComment(
			ctx,
			webhookData.GitLab.ProjectID,
			strconv.Itoa(webhookData.MergeRequest.IID),
			markdownReport,
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
