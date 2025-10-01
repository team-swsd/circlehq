package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/team-swsd/circlehq/internal/log"
	"github.com/team-swsd/circlehq/internal/model"
)

type RouterOptions struct {
	Logger *slog.Logger
}

func DefaultRouter() chi.Router {
	r := chi.NewRouter()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeErrorResponse(w, &model.ErrNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeErrorResponse(w, &model.ErrMethodNotAllowed)
	})
	return r
}

func DefaultRouterOptions(o RouterOptions) ChiServerOptions {
	opts := ChiServerOptions{
		BaseRouter: DefaultRouter(),
		BaseURL:    "",
		Middlewares: []MiddlewareFunc{
			log.AccessHTTPMiddleware(o.Logger),
			log.TraceIDMiddleware(),
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {},
	}
	return opts
}
