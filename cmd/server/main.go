package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/bootstrap"
	"sonarbridge-go/internal/cli/serve"
	"sonarbridge-go/internal/infra/entrypoints/rest"
	"time"

	"github.com/spf13/cobra"
)

const KeyServerAddr = "KeyAddr"

var rootCmd = &cobra.Command{
	Use: "sonarbridge",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serve.NewCommand(func() {

		cfg := configs.Load()

		svc := bootstrap.NewApp(*cfg)

		healthHandler := rest.NewHealthHandler()
		webhookHandler := rest.NewWebhookHandler(svc.Service)
		reportHandler := rest.NewReportHandler(svc.Service)

		router := rest.NewRouter(
			*cfg,
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
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      *router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 20 * time.Second,
			IdleTimeout:  120 * time.Second,
			BaseContext: func(listener net.Listener) context.Context {
				ctx = context.WithValue(ctx, KeyServerAddr, listener.Addr().String())
				return ctx
			},
		}

		/*serverDocs := &http.Server{
			Addr: ":4044",
			Handler: middleware2.NewBuilder(muxDocs).
				Add(middleware2.HeaderAppInfo).
				Add(middleware2.SecurityHeadersMiddleware).
				Add(middleware2.LoggingRequestMiddleware).
				Add(middleware2.RateLimite).
				Build(),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 20 * time.Second,
			IdleTimeout:  120 * time.Second,
			BaseContext: func(listener net.Listener) context.Context {
				ctx = context.WithValue(ctx, KeyServerAddr, listener.Addr().String())
				return ctx
			},
		}*/

		go func() {
			err := serverOne.ListenAndServe()

			if errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Server One closed\n")
			} else if err != nil {
				fmt.Printf("Error listening for server one: %s\n", err)
				os.Exit(1)
			}

			defer cancelCtx()
		}()

		/*go func() {
			err := serverDocs.ListenAndServe()

			if errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Server Docs closed\n")
			} else if err != nil {
				fmt.Printf("Error listening for server docs: %s\n", err)
			}
			defer cancelCtx()
		}()*/

		/*quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		fmt.Println("Server is shutting down...")

		if err := serverOne.Shutdown(ctx); err != nil {
			fmt.Printf("Server forced to shutdown: %s\n", err)
		}*/

		<-ctx.Done()
	}))
	// rootCmd.AddCommand(build.Context)
}

func main() {

	Execute()

}
