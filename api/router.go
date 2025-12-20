package api

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
)

func App(
	routerManager *routerManager,
	mw *middleware.Middleware,
	cfg *structs.Config,
) http.Handler {
	r := chi.NewRouter()

	// Core infra
	r.Use(chiware.RequestID)
	r.Use(chiware.RealIP)
	r.Use(chiware.Recoverer)

	// Limits & security
	r.Use(mw.BodyLimit(int64(cfg.Server.MaxHeaderBytes)))
	r.Use(mw.SecurityHeaders())
	r.Use(mw.RateLimitMiddleware())

	// Observability
	r.Use(mw.SetupLoggerMiddleware())

	// CORS (must be before auth / csrf)
	r.Use(mw.SetupCORS().Handler)

	// Register all routes
	routerManager.RegisterRoutes(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		gecho.Success(w,
			gecho.WithMessage("Welcome to the Mamabloemetjes API"),
			gecho.Send(),
		)
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		gecho.NotFound(w,
			gecho.Send(),
		)
	})

	return r
}
