package auth

import (
	"net/http"

	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/lib"
)

func (ar *AuthRoutesManager) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var blacklistErr error

	if accessToken, err := lib.GetCookieValue(lib.AccessCookieName, r); err == nil {
		claims, err := lib.ParseToken(accessToken, true, ar.cfg.Auth.AccessTokenSecret)
		if err != nil {
			ar.logger.Warn("Failed to parse access token during logout", gecho.Field("error", err))
		} else if err := ar.cacheService.BlacklistToken(claims.Jti, claims.Exp); err != nil {
			ar.logger.Error("Failed to blacklist access token during logout", gecho.Field("error", err))
			blacklistErr = err
		} else if err := ar.cacheService.DeleteUserFromCache(claims.Sub); err != nil {
			ar.logger.Error("Failed to clear user cache during logout",
				gecho.Field("error", err),
				gecho.Field("user_id", claims.Sub),
			)
		}
	}

	if refreshToken, err := lib.GetCookieValue(lib.RefreshCookieName, r); err == nil {
		claims, err := lib.ParseToken(refreshToken, false, ar.cfg.Auth.RefreshTokenSecret)
		if err != nil {
			ar.logger.Warn("Failed to parse refresh token during logout", gecho.Field("error", err))
		} else if err := ar.cacheService.BlacklistToken(claims.Jti, claims.Exp); err != nil {
			ar.logger.Error("Failed to blacklist refresh token during logout", gecho.Field("error", err))
			blacklistErr = err
		}
	}

	if blacklistErr != nil {
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.auth.failedToLogout"),
		))
		return
	}

	lib.ClearCookie(lib.AccessCookieName, w)
	lib.ClearCookie(lib.RefreshCookieName, w)
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.auth.loggedOut"),
	))
}
