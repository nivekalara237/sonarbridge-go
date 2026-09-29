package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/bootstrap"
	"sonarbridge-go/internal/build"
	"sonarbridge-go/internal/cli/server"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/rest"
	"sonarbridge-go/internal/infra/repository/report"
	"sonarbridge-go/internal/logging"
	stringify "sonarbridge-go/pkg/string"
	"strconv"
	"time"
)

const banner = `
  ____   ___  _   _    _    ____      ____  ____   ___ _____   ____   ___   _____ 
 / ___| / _ \| \ | |  / \  |  _ \    | __ )|  _ \_ _|  __ \ | |  _ \ / _ \ | ____|
 \___ \| | | |  \| | / _ \ | |_) |   |  _ \| |_) | || |__) || | |_) | | | ||  _|  
  ___) | |_| | |\  |/ ___ \|  _ <    | |_) |  __/| ||  _  / | |  _ <| |_| || |___ 
 |____/ \___/|_| \_/_/   \_\_| \_\   |____/|_|  |___|_| \_\ | |_| \_\\___/ |_____|
                                                                                   
SONARBRIDE-GO :: Application Started :: Go`

const KeyServerAddr = "KeyAddr"

var rootCmd = server.New()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddFunctionalCommand(func(callableArgs ...any) {
		printBanner()

		// var c = configs.AppConfig
		var c = callableArgs[0].(configs.AppUntypedConfig)

		fmt.Println("++++++++++++++++++")
		fmt.Println(stringify.ToJSON(c))
		// fmt.Println(stringify.ToJSON(callableArgs))
		fmt.Println("++++++++++++++++++")

		if err := logging.InitServer(c.Logging.Level); err != nil {
			log.Fatal(err)
		}

		app := bootstrap.NewApp()

		/*er := app.InitPlugins()
		if er != nil {
			fmt.Println("///////  Error Plugin ////////")
			fmt.Println(er)
			fmt.Println("/////// ////////////// ////////")
		}*/
		healthHandler := rest.NewHealthHandler()
		webhookHandler := rest.NewWebhookHandler(app.Service)
		reportHandler := rest.NewReportHandler(usecase.NewReportService(report.NewRepository()))

		router := rest.NewRouter(
			healthHandler,
			webhookHandler,
			reportHandler,
		)

		muxDocs := http.NewServeMux()

		muxDocs.HandleFunc("/docs", func(writer http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			fmt.Printf("%s: got /docs requests!\n", ctx.Value(KeyServerAddr))
			io.WriteString(writer, "<h1>Documentation</h1>")
		})

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

		<-ctx.Done()
	})
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
