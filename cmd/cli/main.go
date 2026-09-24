package main

import (
	"errors"
	"fmt"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/cli"
	"sonarbridge-go/internal/logging"
)

func main() {
	cfg := configs.Config{LogLevel: "error"}
	if err := logging.InitCLI(cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := cli.NewRootCommand()
	err := root.Execute()
	if err == nil {
		os.Exit(0)
		return
	}

	if cliErr, ok := errors.AsType[*cli.CLIError](err); ok {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", cliErr)
		if cliErr.ShowUsage {
			_, _ = fmt.Fprintln(os.Stderr)
			if cliErr.Command != nil {
				_ = cliErr.Command.Usage()
			} else {
				_ = root.Usage()
			}
		}
		os.Exit(1)
	}

	_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
