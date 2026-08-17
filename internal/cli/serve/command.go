package serve

import (
	"sonarbridge-go/configs"
	"strconv"

	"github.com/spf13/cobra"
)

func NewServeCommand(cfg *configs.Config, mainFunc func(port int, host string)) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:     "serve",
		Short:   "Satrt the HTTP Server",
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

	rootCmd.Flags().String("port", cfg.Port, "Port to listen on")
	rootCmd.Flags().String("host", cfg.Host, "Host to launch on")
	rootCmd.Flags().String("env", cfg.Env, "The environment to run on")

	return rootCmd
}
