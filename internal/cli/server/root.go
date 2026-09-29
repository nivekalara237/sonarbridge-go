package server

import (
	"fmt"
	"os"
	"sonarbridge-go/configs"

	"github.com/spf13/cobra"
)

type RootCmd struct {
}

var rootCmd = &cobra.Command{
	Use:   configs.AppName,
	Short: "CI Bridge Server",
}

func (RootCmd) Execute() error {
	return rootCmd.Execute()
}

func (c RootCmd) AddFunctionalCommand(fn func(callableArgs ...any)) {
	rootCmd.AddCommand(NewServeCommand(fn))
}
func New() *RootCmd {
	return &RootCmd{}
}

var configFile string

func init() {
	fmt.Println("=========== Inside Root ========")
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config",
		"c", "",
		"Path to configuration file: E.g: config.toml, app.yaml, https://my.company.com/cnf/external.json")

	if err := configs.Vconfig.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		panic(err)
	}
	if err := configs.Vconfig.BindPFlag("config", rootCmd.PersistentFlags().ShorthandLookup("c")); err != nil {
		panic(err)
	}
}

func initConfig() {
	cfgFile := configs.Vconfig.GetString("config")

	if cfgFile == "" {
		return
	}

	configs.Vconfig.SetConfigFile(cfgFile)

	if err := configs.Vconfig.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading config file: %v\n", err)
		os.Exit(1)
	}
}
