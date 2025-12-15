package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear access token cookie
	lib.ClearCookie(lib.AccessCookieName, w)
	// Clear refresh token cookie
	lib.ClearCookie(lib.RefreshCookieName, w)

	gecho.Success(w,
		gecho.WithMessage("Logged out successfully"),
		gecho.Send(),
	)
}
