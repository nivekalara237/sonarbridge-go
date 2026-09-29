package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/httpclient"
	"sonarbridge-go/plugins/vcs/gitlab/build"
	gitlabconfig "sonarbridge-go/plugins/vcs/gitlab/config"
	"sonarbridge-go/plugins/vcs/gitlab/service"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/capability"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var logger = hclog.New(&hclog.LoggerOptions{
	Output:     os.Stderr,
	Level:      hclog.Debug, // jamais stdout
	JSONFormat: true,
	Name:       "Gitlab-Plugin",
})

type infoServer struct {
	pluginv1.UnimplementedPluginInfoServer
}

func (s *infoServer) GetInfo(ctx context.Context, req *pluginv1.GetInfoRequest) (*pluginv1.InfoResponse, error) {
	return &pluginv1.InfoResponse{
		Name:            build.Name,
		Version:         build.Version,
		PluginType:      "vcs",
		ProtocolVersion: build.ProtocolVersion,
		Capabilities: []string{
			capability.PullRequestCreateComment,
			capability.PullRequestDeleteComment,
			capability.PullRequestUpdateComment,
			capability.Issue,
			capability.Webhook,
			capability.Artifact,
		},
	}, nil
}

var rootCommand = &cobra.Command{
	Use:           "gitlab-vcs [-c config.yaml] start",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func main() {
	logging.SetOutput(os.Stderr)
	slogLogger := logging.NewLogger(logger)
	slog.SetDefault(slogLogger)

	logFile, err := os.OpenFile("gitlab-plugin.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error("unable to create or open gitlab-plugin.log file: %w", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	multi := io.MultiWriter(os.Stderr, logFile)
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
	logger.Info("GitLab plugin starting")
	config, err := gitlabconfig.LoadConfigs()
	if err != nil {
		return fmt.Errorf("load gitlab config: %w", err)
	}

	var options []httpclient.Option
	if config.CACertPath != "" {
		options = append(options, httpclient.WithTLSCACertFile(config.CACertPath))
	}
	options = append(options, httpclient.WithTimeout(30*time.Second))
	options = append(options, httpclient.WithBaseURL(strings.TrimRight(config.BaseUrl, "/")))
	options = append(options, httpclient.WithRetry(3, 500*time.Millisecond))
	options = append(options, httpclient.WithCache(httpclient.NewInMemoryCache()))
	options = append(options, httpclient.WithDefaultHeader("PRIVATE-TOKEN", config.Token))
	client := httpclient.NewClientHttp()
	commentService := service.NewCommentAndNoteService(client)
	pullrequestService := service.NewPullrequestService(client)
	serveConfig := &goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		GRPCServer:      grpcServer,
		Plugins: plugin.PluginServer(
			plugin.PServerEntry{Key: plugin.PluginKey, ServerImpl: &plugin.InfoGRPCPlugin{Impl: &infoServer{}}},
			plugin.PServerEntry{Key: plugin.CommentAndNoteKey, ServerImpl: &plugin.CommentAndNoteGRPCPlugin{Impl: commentService}},
			plugin.PServerEntry{Key: plugin.PullrequestKey, ServerImpl: &plugin.PullrequestGRPCPlugin{Impl: pullrequestService}},
		),
		Logger: logger,
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
