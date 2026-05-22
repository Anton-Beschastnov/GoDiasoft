package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/api"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	server  *http.Server
	handler api.ServerInterface
	logger  logger.Iface
	host    string
	port    int
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

	// Добавляем middleware для логирования
	r.Use(s.loggingMiddleware)

	// --- ИСПРАВЛЕННАЯ ЛОГИКА SWAGGER ---
	// Мы будем раздавать содержимое папки 'swagger' как статичные файлы.
	// Это включает в себя и `index.html` от Swagger UI, и сам `doc.json`.
	swaggerDir := http.Dir("./swagger/")
	fileServer := http.FileServer(swaggerDir)
	r.Handle("/swagger/*", http.StripPrefix("/swagger/", fileServer))
	// ------------------------------------

	// API хендлеры
	apiHandler := api.HandlerWithOptions(s.handler, api.ChiServerOptions{
		BaseRouter: r,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			s.logger.Error("API error", "path", r.URL.Path, "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	})

	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	s.server = &http.Server{
		Addr:              addr,
		Handler:           apiHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.logger.Info("starting http server", "addr", addr, "swagger_path", "/swagger/index.html")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// loggingMiddleware логирует каждый входящий запрос
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		s.logger.Info("request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
			"status", ww.Status(),
			"size", ww.BytesWritten(),
		)
	})
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping http server")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
