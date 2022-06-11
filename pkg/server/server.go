package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	http.Server
	logger *zap.Logger
	health map[string]bool
	m      *sync.Mutex
}

// New returns a new server
func New(logger *zap.Logger, addr string) *Server {
	s := &Server{
		Server: http.Server{
			Addr: addr,
		},
		logger: logger,
		health: map[string]bool{},
		m:      &sync.Mutex{},
	}
	// routing
	http.HandleFunc("/healthz", s.LoggerHandlerFunc(s.HealthHandler))
	http.HandleFunc("/ready", s.LoggerHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	}))
	http.Handle("/metrics", s.LoggerHandlerFunc(promhttp.Handler().ServeHTTP))

	return s
}

// Run runs the server
func (s *Server) Run() error {
	return s.ListenAndServe()
}
