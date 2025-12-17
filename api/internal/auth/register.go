package auth

import (
	"errors"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.RegisterRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("Please check your registration information and try again"), gecho.Send())
		return
	}

	if body.Email == "" || body.Password == "" || body.Username == "" {
		ar.logger.Warn("Missing required fields in registration", gecho.Field("body", body))
		gecho.BadRequest(w, gecho.WithMessage("Email, username, and password are required"), gecho.Send())
		return
	}

	user, err := ar.authService.Register(body)
	if err != nil {
		if errors.Is(err, lib.ErrConflict) {
			ar.logger.Warn("User already exists", gecho.Field("email", body.Email), gecho.Field("username", body.Username))
			gecho.Conflict(w, gecho.WithMessage("This email or username is already registered"), gecho.Send())
			return
		}
		ar.logger.Error("Failed to create user", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to create account. Please try again"), gecho.Send())
		return
	}

	accessToken, err := ar.authService.GenerateAccessToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate access token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to complete registration. Please try again"), gecho.Send())
		return
	}

	refreshToken, err := ar.authService.GenerateRefreshToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate refresh token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to complete registration. Please try again"), gecho.Send())
		return
	}

	lib.SetCookie(lib.RefreshCookieName, refreshToken, ar.authService.GetRefreshTokenExpiration(), w)
	lib.SetCookie(lib.AccessCookieName, accessToken, ar.authService.GetAccessTokenExpiration(), w)

	// clear password from user
	user.PasswordHash = ""

	gecho.Success(w,
		gecho.WithMessage("User registered successfully"),
		gecho.WithData(user),
		gecho.Send(),
	)
}
