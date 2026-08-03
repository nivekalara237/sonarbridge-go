package serve

import (
	"fmt"
	"sonarbridge-go/configs"

	"github.com/spf13/cobra"
)

func NewCommand(mainFunc func()) *cobra.Command {
	cfg := configs.Load()

	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Run the HTTP Server",
		Aliases: []string{"run", "start"},
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			env, _ := cmd.Flags().GetString("env")
			fmt.Printf("Serving on :%d", port)
			fmt.Printf(". For %s", env)
			// fmt.Println("")
			mainFunc()
			return nil
		},
	}

	cmd.Flags().String("port", cfg.Port, "Port to listen on")
	cmd.Flags().String("env", cfg.Env, "The environment to run on")

	return cmd
}
