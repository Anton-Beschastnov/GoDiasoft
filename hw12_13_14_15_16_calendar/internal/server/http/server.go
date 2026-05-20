package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	logger Logger
	app    Application
	host   string
	port   int
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Application interface{}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	return &Server{
		logger: logger,
		app:    app,
		host:   host,
		port:   port,
	}
}

func (s *Server) Start(_ context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.helloHandler)
	mux.HandleFunc("/hello", s.helloHandler)

	handler := s.loggingMiddleware(mux)

	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	s.server = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.logger.Info("starting http server", "addr", addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping http server")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, Calendar!"))
}
