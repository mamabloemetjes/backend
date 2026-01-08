package auth

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type AuthRoutesManager struct {
	logger       *gecho.Logger
	authService  *services.AuthService
	cacheService *services.CacheService
	emailService *services.EmailService
	orderService *services.OrderService
	cfg          *structs.Config
	mw           *middleware.Middleware
}

func NewAuthRoutesManager(
	logger *gecho.Logger,
	authService *services.AuthService,
	emailService *services.EmailService,
	cacheService *services.CacheService,
	orderService *services.OrderService,
	cfg *structs.Config,
	mw *middleware.Middleware,
) *AuthRoutesManager {
	return &AuthRoutesManager{
		logger:       logger,
		authService:  authService,
		emailService: emailService,
		cacheService: cacheService,
		orderService: orderService,
		cfg:          cfg,
		mw:           mw,
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
		r.Get("/verify-email", rrm.HandleVerifyEmail)

		// Protected routes for user data
		r.Group(func(r chi.Router) {
			r.Use(rrm.mw.UserAuthMiddleware)
			r.Get("/addresses", rrm.HandleGetAddresses)
		})
	})
}
