package http

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type loggerKey struct{}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	log, ok := ctx.Value(loggerKey{}).(*zap.Logger)
	if ok {
		return log
	}

	return zap.L()
}

func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), loggerKey{}, logger)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
