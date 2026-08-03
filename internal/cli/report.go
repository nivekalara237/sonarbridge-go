package cli

import (
	"fmt"
	"sonarbridge-go/internal/core/usecase"

	"github.com/spf13/cobra"
)

func NewReportCommand(svc *usecase.Service, cliCtx *Context) *cobra.Command {
	reportCommand := &cobra.Command{
		Use:   "report",
		Short: "Display SonarQube Analysis as --format=html|md|text|json|yaml|xml",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Executing analyse command...")
			// fmt.Println(options)
			return nil
		},
	}

	reportCommand.Flags().String("format", "json", "The output formatted")

	return reportCommand
}
