package serve

import (
	"sonarbridge-go/configs"
	"strconv"

	"github.com/spf13/cobra"
)

func NewServeCommand(mainFunc func(port int, host string)) *cobra.Command {
	cfg := configs.Load()

	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Run the HTTP Server",
		Aliases: []string{"run", "start"},
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetString("port")
			// env, _ := cmd.Flags().GetString("env")
			addr, _ := cmd.Flags().GetString("host")
			p, _ := strconv.Atoi(port)
			mainFunc(p, addr)
			return nil
		},
	}

	cmd.Flags().String("port", cfg.Port, "Port to listen on")
	cmd.Flags().String("host", cfg.Host, "Host to launch on")
	cmd.Flags().String("env", cfg.Env, "The environment to run on")

	return cmd
}
