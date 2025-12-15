package auth

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type AuthRoutesManager struct {
	logger      *gecho.Logger
	db          *database.DB
	authService *services.AuthService
	cfg         *structs.Config
}

func NewAuthRoutesManager(cfg *structs.Config, logger *gecho.Logger, db *database.DB) *AuthRoutesManager {
	return &AuthRoutesManager{
		logger:      logger,
		db:          db,
		authService: services.NewAuthService(cfg, logger, db),
		cfg:         cfg,
	}
}

func (rrm *AuthRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", rrm.HandleRegister)
		r.Post("/login", rrm.HandleLogin)
		r.Post("/refresh", rrm.HandleRefreshAccessToken)
		r.Post("/logout", rrm.HandleLogout)
		r.Get("/me", rrm.HandleMe)
	})
}
