package cli

import (
	"fmt"
	usecase2 "sonarbridge-go/internal/cli/usecase"
	"strconv"

	"github.com/spf13/cobra"
)

type AnalyseOptions struct {
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
  sonarbridge-cli analyze my-project main --sonar-task-id=AYx123456 --sonar-project-key=my-quality-project --sever-shared-key=xxxx

  # Analyze a Merge Request
  sonarbridge-cli analyze my-project feature/login \
    --mr-iid=42 \
    --commit-sha=abc123 \
    --commit-url=https://gitlab.example.com/... \
    --ci-token=$CI_JOB_TOKEN \
    --sever-shared-key=$CI_JOB_TOKEN \
    --server-url=https://sonarbridge.cavom.lan \
    --server-cert=string-certificate.pem \
    --server-key=string-certificate.key \
    --sonar-task-id=AYx123456 \
    --sonar-project-key=my-quality-project \
    --status=SUCCESS`
	helpLongDescription = `Runs a SonarQube analysis for the specified project and branch.

The project ID and branch are required positional arguments.
Additional metadata can be provided through flags (Merge Request, commit, CI, etc.).`
)

var options AnalyseOptions

func NewAnalyzeCommand(cliCtx *CliContext) *cobra.Command {
	analyzeCommand := &cobra.Command{
		Use:     "analyze  <gitlab-project-id> <branch-name>",
		Short:   "Fetch SonarQube Analysis and create a MR/PR rapport message",
		Long:    helpLongDescription,
		Example: helpExample,
		// Args:    cobra.ExactArgs(2),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return &CLIError{
					Err: fmt.Errorf(
						"accepts 2 args(s), received %d",
						len(args),
					),
					ShowUsage: true,
					Command:   cmd,
				}
			}
			return nil
		},
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

			mr, _ := strconv.Atoi(options.mrIid)

			payload := usecase2.GenerateSonarReportRequest{
				TaskId: options.taskId,
				SonarProject: struct {
					Key string `json:"key"`
				}{Key: options.sonarProjectKey},
				Gitlab: struct {
					ProjectId  string `json:"projectId"`
					CiToken    string `json:"ciToken"`
					BranchName string `json:"branchName"`
					BranchUrl  string `json:"branchUrl"`
					CommitSha  string `json:"commitSha"`
				}{
					ProjectId:  positionArgs.gitlabProjectId,
					BranchName: positionArgs.branchName,
					CommitSha:  options.commitSha,
					BranchUrl:  options.commitUrl,
					CiToken:    options.ciToken,
				},
				MergeRequest: struct {
					IID int `json:"iid"`
				}{IID: mr},
			}

			uc := usecase2.NewGenerateSonarReport(
				cliCtx.serverUrl,
			)

			return uc.ExecuteSonarRequest(
				payload,
				cliCtx.serverSharedKey,
			)
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
	_ = analyzeCommand.MarkFlagRequired("sever-shared-key")

	return analyzeCommand
}
