package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/api"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	server  *http.Server
	handler api.ServerInterface
	logger  logger.Iface
	host    string
	port    int
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func NewServer(logger logger.Iface, handler api.ServerInterface, host string, port int) *Server {
	return &Server{
		handler: handler,
		logger:  logger,
		host:    host,
		port:    port,
	}
}

func (s *Server) Start(_ context.Context) error {
	r := chi.NewRouter()

	// Определяем путь к swagger директории относительно текущей рабочей директории
	cwd, err := os.Getwd()
	if err != nil {
		s.logger.Error("failed to get current working directory", "error", err)
	}
	swaggerDir := filepath.Join(cwd, "swagger")

	// Обслуживаем swagger/doc.json как отдельный файл
	r.Handle("/swagger/doc.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(swaggerDir, "doc.json"))
	}))

	// Swagger UI - используем файлы для обслуживания swagger.json
	swaggerURL := fmt.Sprintf("http://%s:%d/swagger/doc.json", s.host, s.port)
	r.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(swaggerURL),
	))

	// Обслуживаем остальные файлы из swagger директории
	r.Handle("/swagger/{*filepath}", http.StripPrefix("/swagger/", http.FileServer(http.Dir(swaggerDir))))

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
