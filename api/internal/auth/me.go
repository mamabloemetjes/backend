package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleMe(w http.ResponseWriter, r *http.Request) {
	accessToken, err := lib.GetCookieValue(lib.AccessCookieName, r)
	if err != nil {
		ar.logger.Warn("Access token not found in cookies", gecho.Field("error", err))
		gecho.Success(w, gecho.WithMessage("Invalid access token"), gecho.Send())
		return
	}

	claims, err := lib.ParseToken(accessToken, true, ar.cfg.Auth.AccessTokenSecret)
	if err != nil {
		ar.logger.Warn("Failed to parse access token", gecho.Field("error", err))
		gecho.Success(w, gecho.WithMessage("Invalid access token"), gecho.Send())
		return
	}

	user, err := ar.authService.GetUserByID(claims.Sub)

	gecho.Success(w,
		gecho.WithData(user),
		gecho.Send(),
	)
}
