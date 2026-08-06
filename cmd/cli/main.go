package main

import (
	"fmt"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/bootstrap"
	"sonarbridge-go/internal/cli"
	"sonarbridge-go/internal/logging"
)

func main() {
	cfg := configs.Load()

	if err := logging.InitServer(*cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app := bootstrap.NewApp(*cfg)
	root := cli.NewRootCommand(app.Service)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
