package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleMe(w http.ResponseWriter, r *http.Request) {
	accessToken, err := lib.GetCookieValue(lib.AccessCookieName, r)
	if err != nil {
		// check if refresh token is present - signal to the frontend to refresh
		refreshToken, refreshErr := lib.GetCookieValue(lib.RefreshCookieName, r)
		if refreshErr != nil {
			gecho.Success(w, gecho.WithMessage("No access token"), gecho.Send())
			return
		}
		// refresh automatically
		authResponse, err := ar.authService.RefreshAccessToken(refreshToken)
		if err != nil {
			ar.logger.Warn("Failed to refresh access token", gecho.Field("error", err))
			gecho.Success(w, gecho.WithMessage("No access token"), gecho.Send())
			return
		}

		// Set new tokens cookie
		lib.SetCookie(lib.AccessCookieName, authResponse.AccessToken, ar.authService.GetAccessTokenExpiration(), w)
		lib.SetCookie(lib.RefreshCookieName, authResponse.RefreshToken, ar.authService.GetRefreshTokenExpiration(), w)

		// return user data
		gecho.Success(w,
			gecho.WithData(authResponse.User),
			gecho.Send(),
		)
		return
	}

	// Parse and validate access token if access token is still active
	claims, err := lib.ParseToken(accessToken, true, ar.cfg.Auth.AccessTokenSecret)
	if err != nil {
		ar.logger.Error("Failed to parse access token", gecho.Field("error", err))
		gecho.Success(w, gecho.WithMessage("Invalid access token"), gecho.Send())
		return
	}

	user, err := ar.authService.GetUserByID(claims.Sub)

	gecho.Success(w,
		gecho.WithData(user),
		gecho.Send(),
	)
}
