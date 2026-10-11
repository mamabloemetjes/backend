package api

import (
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"

	"github.com/mamabloemetjes/backend/api/middleware"
	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/structs"
)

func App(
	routerManager *routerManager,
	mw *middleware.Middleware,
	cfg *structs.Config,
) http.Handler {
	r := chi.NewRouter()

	// Core infra
	r.Use(chiware.RequestID)
	r.Use(requestIDHeader)
	r.Use(chiware.RealIP)
	r.Use(chiware.Recoverer)

	// Limits & security
	r.Use(mw.BodyLimit(int64(cfg.Server.MaxHeaderBytes)))
	r.Use(mw.SecurityHeaders())
	r.Use(mw.RateLimitMiddleware())

	// Observability
	r.Use(mw.SetupLoggerMiddleware())
	r.Use(middleware.MetricsMiddleware)

	// CORS (must be before auth / csrf)
	r.Use(mw.SetupCORS().Handler)

	// Register all routes
	routerManager.RegisterRoutes(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.Send(gecho.Success(w,
			gecho.WithMessage("success.api.welcome"),
		))
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		response.Send(gecho.NotFound(w))
	})

	return r
}

func requestIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID := chiware.GetReqID(r.Context()); requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
		}
		next.ServeHTTP(w, r)
	})
}
