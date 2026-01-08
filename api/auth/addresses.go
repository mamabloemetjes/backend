package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleGetAddresses(w http.ResponseWriter, r *http.Request) {
	// Get user ID from claims (set by UserAuthMiddleware)
	claims, err := lib.ExtractClaims(r)
	if err != nil {
		ar.logger.Error("Failed to extract claims", gecho.Field("error", err))
		gecho.Unauthorized(w, gecho.WithMessage("error.auth.unauthorized"), gecho.Send())
		return
	}

	// Get all addresses for the user
	addresses, err := ar.orderService.GetUserAddresses(r.Context(), claims.Sub)
	if err != nil {
		ar.logger.Error("Failed to get user addresses",
			gecho.Field("error", err),
			gecho.Field("user_id", claims.Sub),
		)
		gecho.InternalServerError(w, gecho.WithMessage("error.addresses.fetchFailed"), gecho.Send())
		return
	}

	// Get user information
	user, err := ar.authService.GetUserByID(claims.Sub)
	if err != nil {
		ar.logger.Error("Failed to get user information",
			gecho.Field("error", err),
			gecho.Field("user_id", claims.Sub),
		)
		gecho.InternalServerError(w, gecho.WithMessage("error.user.fetchFailed"), gecho.Send())
		return
	}

	// Return user info and addresses
	gecho.Success(w,
		gecho.WithData(map[string]interface{}{
			"user":      user,
			"addresses": addresses,
		}),
		gecho.WithMessage("success.addresses.fetched"),
		gecho.Send(),
	)
}
