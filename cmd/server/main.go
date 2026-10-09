package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/bootstrap"
	"sonarbridge-go/internal/build"
	"sonarbridge-go/internal/cli/server"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/rest"
	"sonarbridge-go/internal/infra/grpc"
	"sonarbridge-go/internal/infra/repository/report"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/provision"
	"strconv"
	"syscall"
	"time"
)

const banner = `

####  ####  ##### ####   #### #####        #### #####
#   # #   #   #   #   # #     #           #       #
####  ####    #   #   # # ### ####  ##### #       #
#   # #  #    #   #   # #   # #           #       #
####  #   # ##### ####   #### #####        #### #####

SONARBRIDE-GO :: Application Started :: Go`

const KeyServerAddr = "KeyAddr"

var rootCmd = server.New()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func desiredOf(vcs []configs.VcsProviderCnf) []provision.Desired {
	ds := make([]provision.Desired, 0, len(vcs))
	for _, v := range vcs {
		ds = append(ds, provision.Desired{
			Name:    v.BinaryName,
			Version: v.Version,
		})
	}
	return ds
}

func init() {
	rootCmd.AddFunctionalCommand(func(callableArgs ...any) {
		printBanner()

		var c = callableArgs[0].(configs.AppUntypedConfig)

		if err := logging.InitServer(c.Logging.Level); err != nil {
			log.Fatal(err)
		}

		app := bootstrap.NewApp()
		app.Bootstrapping()

		healthHandler := rest.NewHealthHandler()
		pluginHealthCheckHandler := rest.NewPluginHealthCheckHandler(
			app.PluginManager,
			desiredOf(configs.AppConfig.VcsProviders),
		)
		webhookHandler := rest.NewWebhookHandler(app.Service)
		vcsHandler := rest.NewPullrequestComment(grpc.NewVcs(app.PluginManager))
		reportHandler := rest.NewReportHandler(usecase.NewReportService(report.NewRepository()))
		pluginCtrler := rest.NewPluginController(app.PluginManager)
		pluginLifecycleCtrler := rest.NewPluginLifecycleController(app.PluginManager)

		router := rest.NewRouter(
			healthHandler,
			webhookHandler,
			reportHandler,
			vcsHandler,
			pluginCtrler,
			pluginLifecycleCtrler,
		)

		muxPluginTelemetry := http.NewServeMux()

		muxPluginTelemetry.HandleFunc(
			configs.AppConfig.Plugins.Telemetry.Liveness,
			pluginHealthCheckHandler.PluginLive,
		)
		muxPluginTelemetry.HandleFunc(
			configs.AppConfig.Plugins.Telemetry.Readiness,
			pluginHealthCheckHandler.PluginReady,
		)

		pluginServer := &http.Server{Handler: muxPluginTelemetry, ReadHeaderTimeout: 10 * time.Second}
		pluginServerErr := make(chan error, 1)

		go func() {
			var listen net.Listener
			var err error
			if listen, err = net.Listen("tcp", fmt.Sprintf("%s:%d",
				configs.AppConfig.Server.Host,
				configs.AppConfig.Plugins.Telemetry.Port)); err != nil {
				pluginServerErr <- fmt.Errorf("plugin/telemetry: listen: %w", err)
				return
			}
			if configs.AppConfig.Server.Tls.Enabled {
				pluginServerErr <- pluginServer.ServeTLS(
					listen,
					configs.AppConfig.Server.Tls.CertFile,
					configs.AppConfig.Server.Tls.KeyFile,
				)
			} else {
				pluginServerErr <- pluginServer.Serve(listen)
			}
		}()

		ctx, cancelCtx := context.WithCancel(context.Background())

		// defer cancelCtx()
		serverOne := &http.Server{
			Addr:              net.JoinHostPort(c.Server.Host, strconv.Itoa(c.Server.Port)),
			Handler:           *router,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      20 * time.Second,
			IdleTimeout:       60 * time.Second,
			TLSConfig:         nil,
			BaseContext: func(listener net.Listener) context.Context {
				ctx = context.WithValue(ctx, KeyServerAddr, listener.Addr().String())
				return ctx
			},
		}

		log.Printf("listening on %s", serverOne.Addr)

		go func() {
			var err error
			if configs.AppConfig.Server.Tls.Enabled {
				err = serverOne.ListenAndServeTLS(
					configs.AppConfig.Server.Tls.CertFile,
					configs.AppConfig.Server.Tls.KeyFile,
				)
			} else {
				err = serverOne.ListenAndServe()
			}

			if errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Server one closed\n")
			} else if err != nil {
				fmt.Printf("Error listening for server one: %s\n", err)
				os.Exit(1)
			}

			defer cancelCtx()
		}()

		logging.Info(
			"http server listening",
			"addr", configs.AppConfig.Server.Host,
			"port", configs.AppConfig.Server.Port,
			// "server/liveness", configs.AppConfig,
			// "server/deadiness", configs.AppConfig,
			"plugin/liveness", configs.AppConfig.Plugins.Telemetry.Liveness,
			"plugin/deadiness", configs.AppConfig.Plugins.Telemetry.Readiness,
		)

		ctxNotify, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		pctx, cancelProvision := context.WithCancel(ctxNotify)
		defer cancelProvision()
		provisioned := make(chan struct{})
		runError := app.RunPluginProvisioner(provisioned, pctx)

		// var runError error
		select {
		case <-ctxNotify.Done():
		case err := <-pluginServerErr:
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					runError = fmt.Errorf("plugin/telemetry: http server: %w", err)
					logging.Error("plugin/telemetry server failed", "err", err)
				}
			}
		}

		log.Println("shutting down")
		// cancelProvision()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := serverOne.Shutdown(shutdownCtx); err != nil && runError == nil {
			runError = fmt.Errorf("app: http shutdown: %w", err)
		} else if errr := pluginServer.Shutdown(shutdownCtx); errr != nil {
			runError = fmt.Errorf("plugin/telemetry/server: http shutdown: %w", errr)
		}

		select {
		case <-provisioned:
		case <-time.After(30 * time.Second):
			logging.Warn("provisioning did not stop in time")
		}
		stopPlugins(app.PluginManager)
		if runError != nil {
			logging.Error("run error", "reason", runError)
		}
	})

}

func stopPlugins(mgr *manager.Manager) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, s := range mgr.List() {
		if s.State != lifecycle.StateReady {
			continue
		}

		if err := mgr.Stop(ctx, s.Name); err != nil {
			logging.Error("plugin did not stop clearly", "plugin", s.Name, "error", err)
		}
	}
}

func main() {
	Execute()
}

func printBanner() {
	fmt.Println(banner)
	fmt.Println()
	fmt.Println(build.GetBuildInfo().ToServerString())
	fmt.Println()
}
