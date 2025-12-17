package api

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/config"
	"mamabloemetjes_server/database"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
)

func App() chi.Router {
	r := chi.NewRouter()

	// create loggers
	logLevel := gecho.ParseLogLevel(config.GetLogLevel())
	mwLogger := gecho.NewLogger(gecho.NewConfig(gecho.WithShowCaller(false), gecho.WithLogLevel(logLevel)))
	standardLogger := gecho.NewLogger(gecho.NewConfig(gecho.WithShowCaller(true), gecho.WithLogLevel(logLevel)))

	// db
	db := database.GetInstance()

	// config
	cfg := config.GetConfig()

	// Initialize middleware
	mw := middleware.NewMiddleware(cfg, mwLogger, db)

	// Core infra
	r.Use(chiware.RequestID)
	r.Use(chiware.RealIP)
	r.Use(chiware.Recoverer)

	// Limits & security
	r.Use(mw.BodyLimit(10 * 1024 * 1024))
	r.Use(mw.SecurityHeaders())

	// Observability
	r.Use(gecho.Handlers.CreateLoggingMiddleware(mwLogger))

	// CORS (must be before auth / csrf)
	r.Use(mw.SetupCORS().Handler)

	// Register all routes
	NewRouterManager(standardLogger, db, cfg, mw).RegisterRoutes(r)

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
