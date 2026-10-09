package local_registry

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type LocalRegistryServer struct {
	port              int
	address           string
	serverMux         *http.ServeMux
	absoluteAssetPath string
}

func NewLocalRegistryServer(port int, addr, absoluteAssetPath string) *LocalRegistryServer {
	return &LocalRegistryServer{
		port:              port,
		address:           addr,
		serverMux:         http.NewServeMux(),
		absoluteAssetPath: absoluteAssetPath,
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri, _ := url.PathUnescape(r.RequestURI)
		fmt.Printf("Request: method=%s, Uri=%s\n", r.Method, uri)
		next.ServeHTTP(w, r)
		log.Println("Request took : " + time.Since(start).String())
	})
}

func (s *LocalRegistryServer) Serve() {
	ctx, cancel := context.WithCancel(context.Background())

	server := http.Server{
		Handler:           logMiddleware(s.serverMux),
		Addr:              net.JoinHostPort(s.address, strconv.Itoa(s.port)),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			// c := listener.
			return ctx
		},
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("Server Error: %s", err)
			fmt.Println()
		}
		defer cancel()
	}()

	fmt.Println("##################################################")
	fmt.Println("##          Local Registry Server               ##")
	fmt.Println("##################################################")
	fmt.Println("")
	fmt.Printf("Listen      : http(s)://%s:%d", s.address, s.port)
	fmt.Println()
	fmt.Println("Start At    : " + time.Now().String())
	fmt.Println("Assets Path : " + s.absoluteAssetPath)
	fmt.Println()

	<-ctx.Done()
}
