package middleware

import (
	"context"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
)

// Context keys for storing user data in request context
type contextKey string

const (
	UserContextKey   contextKey = "user"
	ClaimsContextKey contextKey = "claims"
)

// UserAuthMiddleware protects routes to only logged-in users
func (mw *Middleware) UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := lib.ExtractClaims(r, mw.authService.GetAccessTokenSecret())
		if err != nil {
			mw.logger.Warn("Failed to extract claims from request", gecho.Field("error", err))
			gecho.Unauthorized(w, gecho.WithMessage("Invalid or missing access token"), gecho.Send())
			return
		}

		// Add user and claims to request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

		// Continue to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminAuthMiddleware protects routes to only admin users
// Must be used after UserAuthMiddleware
func (mw *Middleware) AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get claims from context
		claims, err := lib.ExtractClaims(r, mw.authService.GetAccessTokenSecret())
		if err != nil {
			mw.logger.Warn("Failed to extract claims from request", gecho.Field("error", err))
			gecho.Forbidden(w, gecho.WithMessage("Access denied"), gecho.Send())
			return
		}

		// Check if user has admin role
		if claims.Role != "admin" {
			mw.logger.Warn("Non-admin user attempted to access admin route", gecho.Field("user_id", claims.Sub), gecho.Field("role", claims.Role))
			gecho.Forbidden(w, gecho.WithMessage("Admin access required"), gecho.Send())
			return
		}

		// User is admin, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext is a helper function to extract the user from request context
func GetUserFromContext(ctx context.Context) (*tables.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*tables.User)
	return user, ok
}

// GetClaimsFromContext is a helper function to extract the claims from request context
func GetClaimsFromContext(ctx context.Context) (*structs.AuthClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*structs.AuthClaims)
	return claims, ok
}
