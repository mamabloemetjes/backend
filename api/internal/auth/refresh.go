package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	// Extract refresh token from cookies
	refreshToken, err := lib.GetCookieValue(lib.RefreshCookieName, r)
	if err != nil {
		ar.logger.Warn("Refresh token not found in cookies", gecho.Field("error", err))
		gecho.Unauthorized(w, gecho.WithMessage("Refresh token missing"), gecho.Send())
		return
	}

	authResponse, err := ar.authService.RefreshAccessToken(refreshToken)
	if err != nil {
		ar.logger.Warn("Failed to refresh access token", gecho.Field("error", err))
		gecho.Unauthorized(w, gecho.WithMessage("Invalid refresh token"), gecho.Send())
		return
	}

	// Set new access token cookie
	lib.SetCookie(lib.AccessCookieName, authResponse.AccessToken, ar.authService.GetAccessTokenExpiration(), w)

	gecho.Success(w,
		gecho.WithMessage("Access token refreshed successfully"),
		gecho.WithData(authResponse.User),
		gecho.Send(),
	)
}
