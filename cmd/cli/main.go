package main

import (
	"fmt"
	"log"
	"os"
	"sonarbridge-go/configs"
)

func main()  {
	cfg := configs.Load()

	fmt.Println("The config are", cfg)

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("	cli analyze <projectKey> <branch>")
		os.Exit(1)
	}

	args := os.Args

	switch args[1] {
	case "analyze":
		if len(args) != 4 {
			log.Fatal("Usage: cli analyze <project> <branch>")
		}

		projectKey := args[2]
		branch := args[3]

		fmt.Println("We analyse for project", projectKey, "on", branch, "branch")
		fmt.Println("Bye!")

	default:
		log.Fatalf("Unknown command: %s", args[1])
	}
}
