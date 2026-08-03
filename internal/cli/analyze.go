package cli

import (
	"context"
	"fmt"
	"log"
	"sonarbridge-go/internal/core/domain"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/utils"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type analyseOptions struct {
	mrIid           string
	commitUrl       string
	commitSha       string
	ciToken         string
	taskId          string
	analysisStatus  string
	sonarProjectKey string
}

type positionalArgs struct {
	gitlabProjectId string
	branchName      string
}

const (
	helpExample = `  # Analyze the main branch (outside of mr/pr)
  sonarbridge-cli analyze my-project main --sonar-task-id=AYx123456 --sonar-project-key=my-quality-project

  # Analyze a Merge Request
  sonarbridge-cli analyze my-project feature/login \
    --mr-iid=42 \
    --commit-sha=abc123 \
    --commit-url=https://gitlab.example.com/... \
    --ci-token=$CI_JOB_TOKEN \
    --sonar-task-id=AYx123456 \
    --sonar-project-key=my-quality-project \
    --status=SUCCESS`
	helpLongDescription = `Runs a SonarQube analysis for the specified project and branch.

The project ID and branch are required positional arguments.
Additional metadata can be provided through flags (Merge Request, commit, CI, etc.).`
)

var options analyseOptions

func NewAnalyzeCommand(svc *usecase.Service, cliCtx *Context) *cobra.Command {
	analyzeCommand := &cobra.Command{
		Use:     "analyze  <project-id> <branch>",
		Short:   "Fetch SonarQube Analysis and créate a MR/PR rapport message",
		Long:    helpLongDescription,
		Example: helpExample,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Executing analyse command...")

			if cliCtx.Verbose {
				fmt.Println("Verbose enabled")

			}
			cliCtx.Output.Debug(cliCtx.Verbose, "")

			positionArgs := positionalArgs{
				gitlabProjectId: args[0],
				branchName:      args[1],
			}

			fmt.Printf("ProjectID=%s and branch=%s\n", positionArgs.gitlabProjectId, positionArgs.branchName)
			fmt.Println(options)

			svcCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			mr, _ := strconv.Atoi(options.mrIid)

			payload := domain.SonarQubeWebhookPayload{
				TaskID: &options.taskId,
				Status: utils.Ternary[domain.WebhookStatus](options.analysisStatus == "SUCCESS", "SUCCESS", "FAILED"),
				SonarProject: domain.Project{
					Key: options.sonarProjectKey,
				},
				GitLab: domain.GitLab{
					ProjectID: positionArgs.gitlabProjectId,
					CIToken:   &options.ciToken,
				},
				Branch: &domain.Branch{
					Name: positionArgs.branchName,
					URL:  &options.commitUrl,
					Commit: &domain.Commit{
						SHA:     options.commitSha,
						Message: "",
					},
				},
				MergeRequest: &domain.MergeRequest{
					IID: mr,
				},
				Properties: nil,
			}

			response, err := svc.Execute(svcCtx, payload)
			if err != nil {
				return err
			}

			log.Println("Response cli:", response.Received, response.Mergeable, response.QualityGateStatus)

			if options.mrIid != "" && (response.Mergeable && response.Received) {
				return fmt.Errorf("something wrong on your code, check report of analysis")
			} else if !response.Received {
				return fmt.Errorf("unknown error")
			}
			return cliCtx.Output.Print(response)
		},
	}

	analyzeCommand.Flags().StringVar(
		&options.mrIid,
		"mr-iid",
		"",
		"The merge-request or pull-request ID if the pipeline is triggered on MR/PR",
	)

	analyzeCommand.Flags().StringVar(
		&options.commitUrl,
		"commit-url",
		"",
		"The full commit url where the job is triggered on",
	)

	analyzeCommand.Flags().StringVar(
		&options.commitSha,
		"commit-sha",
		"",
		"The full commit SHA where the job is triggered on",
	)

	analyzeCommand.Flags().StringVar(
		&options.ciToken,
		"ci-token",
		"",
		"The ci-token",
	)

	analyzeCommand.Flags().StringVar(
		&options.taskId,
		"sonar-task-id",
		"",
		"The sonar task ID",
	)

	analyzeCommand.Flags().StringVar(
		&options.analysisStatus,
		"status",
		"OK",
		"The sonar task analysis status",
	)

	analyzeCommand.Flags().StringVar(
		&options.sonarProjectKey,
		"sonar-project-key",
		"",
		"The sonar project key",
	)

	_ = analyzeCommand.MarkFlagRequired("sonar-project-key")
	_ = analyzeCommand.MarkFlagRequired("sonar-task-id")

	return analyzeCommand
}
