package cli

import (
	"fmt"
	"sonarbridge-go/internal/build"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	info := build.GetBuildInfo()
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", info.Commit)
			fmt.Printf("Commit: %s\n", info.Commit)
			fmt.Printf("Built: %s\n", info.Date)
			fmt.Println(info.ToCliString())
		},
	}
}
