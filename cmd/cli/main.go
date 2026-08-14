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
	// cfg := configs.Load()

	cfg := configs.Config{LogLevel: "error"}

	if err := logging.InitCLI(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// app := bootstrap.NewApp(*cfg)
	root := cli.NewRootCommand()

	err := root.Execute()

	if err == nil {
		os.Exit(0)
		return
	}

	var cliErr *cli.CLIError

	if errors.As(err, &cliErr) {
		fmt.Fprintln(os.Stderr, "Error:", cliErr)

		if cliErr.ShowUsage {
			fmt.Fprintln(os.Stderr)
			if cliErr.Command != nil {
				_ = cliErr.Command.Usage()
			} else {
				_ = root.Usage()
			}
		}
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
