package httpserver

import (
	"context"
	"log"
	"net/http"
)

type MyServer struct {
	HTTP *http.Server
}

func NewServer(addr string, router http.Handler) *MyServer {
	return &MyServer{HTTP: &http.Server{Addr: addr, Handler: router}}
}

func (srv *MyServer) Start() {
	log.Println("Starting http-server on addr:\t", srv.HTTP.Addr)
	err := srv.HTTP.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting http server on %s -\t%v", srv.HTTP.Addr, err)
	}
}

func (srv *MyServer) Stop(ctx context.Context) error {
	log.Println("Shutting down http-server")

	return srv.HTTP.Shutdown(ctx)
}
