package server

import (
	"sonarbridge-go/configs"

	"github.com/spf13/cobra"
)

func runServer(mainFunc func(callableArgs ...any)) error {

	configs.InitConfig()
	configs.SetDefaults()

	if err := configs.LoadConfigFile(); err != nil {
		return err
	}

	cfg, err := configs.LoadAppConfig()
	if err != nil {
		return err
	}
	mainFunc(cfg)
	return nil
}

func init() {
	// flags := serveCmd.Flags()
	// configs.BindServeFlags(flags)
	// rootCmd.AddCommand(serveCmd)
}

func NewServeCommand(mainFunc func(callableArgs ...any)) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start the HTTP Server",
		Aliases: []string{"run", "start"},
		RunE: func(cmd *cobra.Command, args []string) error {
			/*port, _ := cmd.Flags().GetString("port")
			addr, _ := cmd.Flags().GetString("host")
			p, _ := strconv.Atoi(port)
			if configFile != "" {
				configs.AppConfig.SetConfigFile(configFile)
				if err := configs.AppConfig.ReadInConfig(); err != nil {
					return fmt.Errorf("failed to read config: %w", err)
				}
			}
			mainFunc(p, addr)
			return nil*/
			return runServer(mainFunc)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return configs.ApplyServeFlags(cmd.Flags())
		},
	}

	return serveCmd
}
