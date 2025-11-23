package router

import "net/http"

type handler interface {
	InjectRoutes(*http.ServeMux)
}

func New(handlers ...handler) *http.ServeMux {
	mux := http.NewServeMux()
	for _, h := range handlers {
		h.InjectRoutes(mux)
	}

	return mux
}
