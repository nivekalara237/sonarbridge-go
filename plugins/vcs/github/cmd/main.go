package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/httpclient"
	"sonarbridge-go/plugins/vcs/github/build"
	githubconfig "sonarbridge-go/plugins/vcs/github/config"
	"sonarbridge-go/plugins/vcs/github/service"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/shared"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var logger hclog.Logger

var rootCommand = &cobra.Command{
	Use:           "github-vcs [-c config.yaml] start",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func main() {

	logFile, err := os.OpenFile("github-plugin.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error("unable to create or open github-plugin.log file: %w", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	multi := io.MultiWriter(os.Stderr, logFile)

	logger = hclog.New(&hclog.LoggerOptions{
		Output:     multi,
		Level:      hclog.Debug, // jamais stdout
		JSONFormat: true,
		Name:       "Github-Plugin",
	})

	logging.SetOutput(multi)
	slogLogger := logging.NewLogger(logger)
	slog.SetDefault(slogLogger)

	log.SetOutput(multi)

	if err := rootCommand.Execute(); err != nil {
		logger.Error("gitlab plugin exited with an error ", "error", err)
		os.Exit(1)
	}
}

func init() {
	rootCommand.AddCommand(&cobra.Command{
		Use:     "start",
		Aliases: []string{"run"},
		RunE:    runStart,
	})
}

func runStart(cmd *cobra.Command, args []string) error {
	logger.Info("GitHub plugin starting")
	config, err := githubconfig.LoadConfigs()
	if err != nil {
		return fmt.Errorf("load github config: %w", err)
	}

	var options []httpclient.Option
	if config.CACertPath != "" {
		options = append(options, httpclient.WithTLSCACertFile(config.CACertPath))
	}
	options = append(options, httpclient.WithTimeout(30*time.Second))
	options = append(options, httpclient.WithBaseURL(strings.TrimRight(config.BaseUrl, "/")))
	options = append(options, httpclient.WithRetry(3, 500*time.Millisecond))
	options = append(options, httpclient.WithCache(httpclient.NewInMemoryCache()))
	options = append(options, httpclient.WithDefaultHeader("ACCESS-TOKEN", config.Token))

	if config.CACertPath != "" {
		options = append(options, httpclient.WithTLSCACertFile(config.CACertPath))
	}

	client := httpclient.NewClientHttp(options...)
	pullrequestService := service.NewPullrequestService(client)

	plugins, err := plugin.PluginServer(
		plugin.Meta{
			Name:            build.Name,
			Version:         build.Version,
			PluginType:      "vcs",
			ProtocolVersion: 1,
		},
		shared.Pullrequest(pullrequestService),
	)

	if err != nil {
		return fmt.Errorf("build plugin: %w", err)
	}

	serveConfig := &goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		GRPCServer:      grpcServer,
		Plugins:         plugins,
		Logger:          logger,
	}
	logger.Debug("serve config:", "grpc_server_nil", serveConfig.GRPCServer == nil)
	goplugin.Serve(serveConfig)

	return nil
}

func grpcServer(ops []grpc.ServerOption) *grpc.Server {
	ops = append(ops,
		grpc.ChainUnaryInterceptor(
			plugin.RecoveryInterceptor(logger),
			plugin.LoggingInterceptor(logger),
			//plugin.AuthInteractor("my-shared-key")
		),
		grpc.ChainStreamInterceptor(
			plugin.StreamLoggingInterceptor(logger)))
	return grpc.NewServer(ops...)
}
