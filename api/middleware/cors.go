package middleware

import (
	"github.com/rs/cors"

	"github.com/mamabloemetjes/backend/config"
)

func (mw *Middleware) SetupCORS() *cors.Cors {
	cfg := config.GetConfig()
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   cfg.Cors.ExposedHeaders,
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	})

	return corsMiddleware
}
