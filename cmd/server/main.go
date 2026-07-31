package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	middleware2 "sonarbridge-go/cmd/server/middleware"
	"sonarbridge-go/configs"
	"sonarbridge-go/internal/core/usecase"
	"sonarbridge-go/internal/infra/entrypoints/rest"
	"sonarbridge-go/internal/infra/http/gitlab"
	"sonarbridge-go/internal/infra/http/sonar"
	"sonarbridge-go/internal/infra/repository"
	"time"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got / request", ctx.Value(KeyServerAddr))
	_, err := io.WriteString(w, "This is my website!\n")
	if err != nil {
		fmt.Println("Error root:", err)
		return
	}
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s: got /hello request\n", ctx.Value(KeyServerAddr))
	io.WriteString(w, "hello, HTTP!\n")
}

func healthCheck(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(`{"status": "healthy"}`))
}

const KeyServerAddr = "KeyAddr"

func main() {

	cfg := configs.Load()

	svc := &rest.UseCase{
		WebhookUseCase: &usecase.Service{
			SonarInteractor:  &sonar.Interactor{},
			GitlabInteractor: &gitlab.Interactor{},
			ReportInteractor: &repository.Interactor{},
		},
	}

	mux := http.NewServeMux()
	muxDocs := http.NewServeMux()

	mux.HandleFunc("/webhook/sonar", svc.Handler)

	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)
	// mux.HandleFunc("/api/message", handleJSON)
	mux.HandleFunc("/healthz", healthCheck)

	muxDocs.HandleFunc("/docs", func(writer http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fmt.Printf("%s: got /docs requests!\n", ctx.Value(KeyServerAddr))
		io.WriteString(writer, "<h1>Documentation</h1>")
	})

	ctx, cancelCtx := context.WithCancel(context.Background())

	// defer cancelCtx()

	serverOne := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware2.NewBuilder(mux).
			Add(func(handler http.Handler) http.Handler {
				return middleware2.Cors(handler, *cfg.Cors)
			}).
			Add(middleware2.HeaderAppInfo).
			Add(middleware2.RateLimite).
			Add(middleware2.SecurityHeadersMiddleware).
			Add(middleware2.LoggingRequestMiddleware).
			Build(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			ctx = context.WithValue(ctx, KeyServerAddr, listener.Addr().String())
			return ctx
		},
	}

	serverDocs := &http.Server{
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
	}

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

	go func() {
		err := serverDocs.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server Docs closed\n")
		} else if err != nil {
			fmt.Printf("Error listening for server docs: %s\n", err)
		}
		defer cancelCtx()
	}()

	/*quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("Server is shutting down...")

	if err := serverOne.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %s\n", err)
	}*/

	<-ctx.Done()

}
