package main

import (
	"fmt"
	"log"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/bootstrap"
	"sonarbridge-go/internal/cli"
)

func main() {
	cfg := configs.Load()
	app := bootstrap.NewApp(*cfg)

	fmt.Println("The config are", cfg)

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("	cli analyze <projectKey> <branch>")
		os.Exit(1)
	}

	root := cli.NewRootCommand(app.Service)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
