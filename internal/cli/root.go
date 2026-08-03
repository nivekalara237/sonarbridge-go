package cli

import (
	"os"
	"sonarbridge-go/internal/core/usecase"

	"github.com/spf13/cobra"
)

func NewRootCommand(svc *usecase.Service) *cobra.Command {

	cliCtx := &Context{}
	var outputFormat string
	rootCommand := &cobra.Command{
		Use:   "sonarbridge-cli",
		Short: "SonarBridge CLI",
	}

	rootCommand.PersistentFlags().BoolVarP(&cliCtx.Verbose, "verbose", "v", false, "Enabled verbose logs")
	rootCommand.PersistentFlags().StringVar(&cliCtx.LogLevel, "log-level", "debug", "Log level(debug, info, warn, error)")
	rootCommand.PersistentFlags().StringVar(&cliCtx.ConfigFile, "config", "", "Configuration file path")
	rootCommand.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (json, table, text)")

	cliCtx.Output = NewOutput(os.Stdout, outputFormat)

	rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return cliCtx.Validate()
	}

	rootCommand.AddCommand(NewAnalyzeCommand(svc, cliCtx))
	rootCommand.AddCommand(NewReportCommand(svc, cliCtx))
	rootCommand.AddCommand(NewVersionCommand())

	return rootCommand
}
