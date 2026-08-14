package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {

	cliCtx := &CliContext{}
	var outputFormat OutputFormat
	rootCommand := &cobra.Command{
		Use:           "sonarbridge-cli",
		Short:         "SonarBridge CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCommand.PersistentFlags().BoolVarP(&cliCtx.Verbose, "verbose", "v", false, "Enabled verbose logs")
	rootCommand.PersistentFlags().StringVar(&cliCtx.LogLevel, "log-level", "debug", "Log level(debug, info, warn, error)")
	rootCommand.PersistentFlags().StringVar(&cliCtx.ConfigFile, "config", "", "Configuration file path")
	rootCommand.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (json, table, text)")
	rootCommand.PersistentFlags().StringVar(&cliCtx.serverSharedKey, "server-shared-key", "", "The server shared key (webhook token)")
	rootCommand.PersistentFlags().StringVar(&cliCtx.serverKey, "server-key", "", "The client certificate key")
	rootCommand.PersistentFlags().StringVar(&cliCtx.serverCert, "server-cert", "", "The client certificate")
	rootCommand.PersistentFlags().StringVar(&cliCtx.serverUrl, "server-url", "", "The client certificate")

	cliCtx.Output = NewOutput(os.Stdout, outputFormat)

	rootCommand.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return cliCtx.Validate()
	}

	rootCommand.AddCommand(NewAnalyzeCommand(cliCtx))
	rootCommand.AddCommand(NewReportCommand(cliCtx))
	rootCommand.AddCommand(NewVersionCommand())

	return rootCommand
}
