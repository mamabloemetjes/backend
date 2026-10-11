package middleware

import (
	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/database"
	"github.com/mamabloemetjes/backend/services"
	"github.com/mamabloemetjes/backend/structs"
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
