package middleware

import (
	"mamabloemetjes_server/config"

	"github.com/go-chi/cors"
)

func (mw *Middleware) SetupCORS() *cors.Cors {
	cfg := config.GetConfig()
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowOrigins,
		AllowedMethods:   cfg.Cors.AllowMethods,
		AllowedHeaders:   cfg.Cors.AllowHeaders,
		ExposedHeaders:   cfg.Cors.ExposedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	})

	return corsMiddleware
}
