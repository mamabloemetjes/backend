package auth

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.AuthRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("Invalid body"), gecho.Send())
		return
	}

	if body.Email == "" || body.Password == "" {
		ar.logger.Warn("Missing required fields in login", gecho.Field("body", body))
		gecho.BadRequest(w, gecho.WithMessage("Email and password are required"), gecho.Send())
		return
	}

	user, err := ar.authService.Login(body)
	if err != nil {
		ar.logger.Warn("Login failed", gecho.Field("error", err))
		gecho.Unauthorized(w, gecho.WithMessage("Invalid credentials"), gecho.Send())
		return
	}

	accessToken, err := ar.authService.GenerateAccessToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate access token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Failed to generate access token"), gecho.Send())
		return
	}

	refreshToken, err := ar.authService.GenerateRefreshToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate refresh token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Failed to generate refresh token"), gecho.Send())
		return
	}

	lib.SetCookie(lib.RefreshCookieName, refreshToken, ar.authService.GetRefreshTokenExpiration(), w)
	lib.SetCookie(lib.AccessCookieName, accessToken, ar.authService.GetAccessTokenExpiration(), w)

	// clear password from user
	user.PasswordHash = ""

	gecho.Success(w,
		gecho.WithMessage("Login successful"),
		gecho.WithData(user),
		gecho.Send(),
	)
}
