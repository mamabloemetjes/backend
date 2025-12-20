package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleLogout(w http.ResponseWriter, r *http.Request) {

	accessToken, err := lib.GetCookieValue(lib.AccessCookieName, r)
	if err != nil {
		gecho.Success(w,
			gecho.WithMessage("No access token found"),
			gecho.Send(),
		)
		return
	}

	claims, err := lib.ParseToken(accessToken, true, ar.cfg.Auth.AccessTokenSecret)
	if err != nil {
		ar.logger.Error("Failed to parse access token during logout", gecho.Field("error", err))
		gecho.Success(w,
			gecho.WithMessage("Invalid access token"),
			gecho.Send(),
		)
		return
	}

	err = ar.cacheService.BlacklistToken(claims.Jti, claims.Exp)
	if err != nil {
		ar.logger.Error("Failed to blacklist access token during logout", gecho.Field("error", err))
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to logout"),
			gecho.Send(),
		)
		return
	}

	// Also clear user from cache
	if err = ar.cacheService.DeleteUserFromCache(claims.Sub); err != nil {
		ar.logger.Error("Failed to clear user cache during logout", gecho.Field("error", err), gecho.Field("user_id", claims.Sub))
	} else {
		ar.logger.Debug("User cache cleared during logout", gecho.Field("user_id", claims.Sub))
	}

	// Clear access token cookie
	lib.ClearCookie(lib.AccessCookieName, w)
	// Clear refresh token cookie
	lib.ClearCookie(lib.RefreshCookieName, w)

	// Blacklist the token

	gecho.Success(w,
		gecho.WithMessage("Logged out successfully"),
		gecho.Send(),
	)
}
