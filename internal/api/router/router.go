package router

import (
	"net/http"

	gh "github.com/gorilla/handlers"
	"github.com/lezzercringe/avito-test-assignment/internal/api/middleware"
	"github.com/lezzercringe/avito-test-assignment/internal/config"
	"go.uber.org/zap"
)

type handler interface {
	InjectRoutes(*http.ServeMux)
}

func stack(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func New(log *zap.Logger, cfg config.Config, handlers ...handler) http.Handler {
	mux := http.NewServeMux()
	for _, h := range handlers {
		h.InjectRoutes(mux)
	}

	return stack(
		mux,
		gh.RecoveryHandler(),
		middleware.Logging(log),
		middleware.Timeout(cfg.RequestTimeout),
	)
}
