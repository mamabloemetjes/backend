package middleware

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
)

type Middleware struct {
	logger       *gecho.Logger
	authService  *services.AuthService
	cacheService *services.CacheService
	cfg          *structs.Config
}

func NewMiddleware(cfg *structs.Config, logger *gecho.Logger, db *database.DB) *Middleware {
	return &Middleware{
		logger:       logger,
		authService:  services.NewAuthService(cfg, logger, db),
		cacheService: services.NewCacheService(logger, cfg),
		cfg:          cfg,
	}
}
