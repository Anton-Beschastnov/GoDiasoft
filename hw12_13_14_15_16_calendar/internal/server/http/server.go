package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/api"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	server  *http.Server
	handler api.ServerInterface
	logger  Logger
	host    string
	port    int
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func NewServer(logger Logger, handler api.ServerInterface, host string, port int) *Server {
	return &Server{
		handler: handler,
		logger:  logger,
		host:    host,
		port:    port,
	}
}

func (s *Server) Start(_ context.Context) error {
	r := chi.NewRouter()

	apiHandler := api.HandlerWithOptions(s.handler, api.ChiServerOptions{
		BaseRouter: r,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			s.logger.Error("API error", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	})

	handler := s.loggingMiddleware(apiHandler)

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
