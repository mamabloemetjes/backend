package auth

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type AuthRoutesManager struct {
	logger       *gecho.Logger
	db           *database.DB
	authService  *services.AuthService
	cacheService *services.CacheService
	cfg          *structs.Config
	mw           *middleware.Middleware
}

func NewAuthRoutesManager(cfg *structs.Config, logger *gecho.Logger, db *database.DB, mw *middleware.Middleware) *AuthRoutesManager {
	return &AuthRoutesManager{
		logger:      logger,
		db:          db,
		authService: services.NewAuthService(cfg, logger, db),
		cacheService: services.NewCacheService(
			logger,
			cfg,
		),
		cfg: cfg,
		mw:  mw,
	}
}

func (rrm *AuthRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		// CSRF token endpoint (must be called before protected routes)
		r.Get("/csrf", rrm.HandleCSRF)

		// Public routes
		r.Group(func(r chi.Router) {
			r.Use(rrm.mw.CSRFMiddleware())
			r.Post("/register", rrm.HandleRegister)
			r.Post("/login", rrm.HandleLogin)
			r.Post("/logout", rrm.HandleLogout)
		})
		r.Get("/me", rrm.HandleMe)
	})
}
