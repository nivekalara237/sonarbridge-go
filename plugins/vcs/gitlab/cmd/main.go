package main

import (
	"context"
	"fmt"
	"os"
	"sonarbridge-go/pkg/httpclient"
	"sonarbridge-go/plugins/vcs/gitlab/build"
	gitlabconfig "sonarbridge-go/plugins/vcs/gitlab/config"
	"sonarbridge-go/plugins/vcs/gitlab/service"
	"strings"
	"time"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/capability"
	"github.com/nivekalara237/ci-bridge-plugin-sdk/plugin"
	pluginv1 "github.com/nivekalara237/ci-bridge-plugin-sdk/plugin/v1"
	"github.com/spf13/cobra"
)

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
	Use: "gitlab-vcs [-c config.yaml] start",
}

func main() {

	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCommand.AddCommand(&cobra.Command{
		Use:     "start",
		Aliases: []string{"run"},
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Println("Here")
			config, err := gitlabconfig.LoadConfigs()
			if err != nil {
				// fmt.Fprintln(os.Stderr, err)
				// os.Exit(1)
				return err
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

			goplugin.Serve(&goplugin.ServeConfig{
				HandshakeConfig: plugin.Handshake,
				GRPCServer:      goplugin.DefaultGRPCServer,
				Plugins: plugin.PluginServer(
					plugin.PServerEntry{Key: plugin.PluginKey, ServerImpl: plugin.InfoGRPCPlugin{Impl: &infoServer{}}},
					plugin.PServerEntry{Key: plugin.CommentAndNoteKey, ServerImpl: plugin.CommentAndNoteGRPCPlugin{Impl: commentService}},
					plugin.PServerEntry{Key: plugin.PullrequestKey, ServerImpl: plugin.PullrequestGRPCPlugin{Impl: pullrequestService}},
				),
			})

			return nil
		}})
}
