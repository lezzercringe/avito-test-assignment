package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logging(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			aw := &augmentedResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(aw, r)

			duration := time.Since(start)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", aw.statusCode),
				zap.Duration("duration", duration),
			)
		})
	}
}

type augmentedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (aw *augmentedResponseWriter) WriteHeader(code int) {
	aw.statusCode = code
	aw.ResponseWriter.WriteHeader(code)
}
