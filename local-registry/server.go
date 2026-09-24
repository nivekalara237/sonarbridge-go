package local_registry

import (
	"context"
	"fmt"
	"net"
	"net/http"
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

func (s *LocalRegistryServer) Serve() {
	ctx, cancel := context.WithCancel(context.Background())

	server := http.Server{
		Handler:           s.serverMux,
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
			fmt.Sprintln("Server Error: %w", err)
		}
		defer cancel()
	}()

	fmt.Println("##################################################")
	fmt.Println("##          Local Registry Server               ##")
	fmt.Println("##################################################")
	fmt.Println("")
	fmt.Printf("Listen      : %s:%d", s.address, s.port)
	fmt.Println()
	fmt.Println("Start At    : " + time.Now().String())
	fmt.Println("Assets Path : " + s.absoluteAssetPath)
	fmt.Println()

	<-ctx.Done()
}
