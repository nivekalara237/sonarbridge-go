package cli

import (
	"fmt"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/logging"

	"github.com/spf13/cobra"
)

func NewReportCommand(svc *usecase.Service, cliCtx *Context) *cobra.Command {
	reportCommand := &cobra.Command{
		Use:   "report",
		Short: "Display SonarQube Analysis as --format=html|md|text|json|yaml|xml",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Executing analyse command...")
			logging.Info("Executing analyse command...")
			logging.Debug("Executing analyse command...")
			logging.Warn("Executing analyse command...")
			logging.Error("Executing analyse command...")
			// fmt.Println(options)
			return nil
		},
	}

	reportCommand.Flags().String("format", "json", "The output formatted")

	return reportCommand
}
