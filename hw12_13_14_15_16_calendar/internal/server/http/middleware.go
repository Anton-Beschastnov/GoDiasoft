package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		latency := time.Since(start)

		clientIP := r.RemoteAddr
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		logMsg := fmt.Sprintf("%s [%s] %s %s %s %d %d %q",
			clientIP,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.RequestURI,
			r.Proto,
			rw.statusCode,
			latency.Milliseconds(),
			userAgent,
		)

		s.logger.Info(logMsg)
	})
}
