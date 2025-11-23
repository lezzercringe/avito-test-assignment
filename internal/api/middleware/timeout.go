package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(t time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if t != 0 {
				ctx, cancel := context.WithTimeout(r.Context(), t)
				defer cancel()
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
