package auth

import (
	"net/http"

	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/lib"
)

func (ar *AuthRoutesManager) HandleMe(w http.ResponseWriter, r *http.Request) {
	accessToken, err := lib.GetCookieValue(lib.AccessCookieName, r)
	if err != nil {
		// check if refresh token is present - signal to the frontend to refresh
		refreshToken, refreshErr := lib.GetCookieValue(lib.RefreshCookieName, r)
		if refreshErr != nil {
			response.Send(gecho.Success(w, gecho.WithMessage("error.auth.noAccessToken")))
			return
		}
		// refresh automatically
		authResponse, err := ar.authService.RefreshAccessToken(refreshToken)
		if err != nil {
			ar.logger.Warn("Failed to refresh access token", gecho.Field("error", err))
			response.Send(gecho.Unauthorized(w, gecho.WithMessage("error.auth.sessionExpired")))
			return
		}

		// Set new tokens cookie
		lib.SetCookie(lib.AccessCookieName, authResponse.AccessToken, ar.authService.GetAccessTokenExpiration(), w)
		lib.SetCookie(lib.RefreshCookieName, authResponse.RefreshToken, ar.authService.GetRefreshTokenExpiration(), w)
		// return user data
		response.Send(gecho.Success(w,
			gecho.WithData(authResponse.User),
		))
		return
	}

	// Parse and validate access token if access token is still active
	claims, err := lib.ParseToken(accessToken, true, ar.cfg.Auth.AccessTokenSecret)
	if err != nil {
		ar.logger.Error("Failed to parse access token", gecho.Field("error", err))
		response.Send(gecho.Unauthorized(w, gecho.WithMessage("error.auth.invalidAccessToken")))
		return
	}

	user, err := ar.authService.GetUserByID(claims.Sub)
	response.Send(gecho.Success(w,
		gecho.WithData(user),
	))
}
