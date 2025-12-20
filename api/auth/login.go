package auth

import (
	"fmt"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.AuthRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("Please check your login information and try again"), gecho.Send())
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

	if !user.EmailVerified {
		ar.logger.Warn("Email not verified", gecho.Field("userID", user.Id))
		gecho.Forbidden(w, gecho.WithMessage(fmt.Sprintf("Please verify your email address (%s) before logging in", user.Email)), gecho.Send())
		return
	}

	accessToken, err := ar.authService.GenerateAccessToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate access token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to complete login. Please try again"), gecho.Send())
		return
	}

	refreshToken, err := ar.authService.GenerateRefreshToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate refresh token", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to complete login. Please try again"), gecho.Send())
		return
	}

	lib.SetCookie(lib.RefreshCookieName, refreshToken, ar.authService.GetRefreshTokenExpiration(), w)
	lib.SetCookie(lib.AccessCookieName, accessToken, ar.authService.GetAccessTokenExpiration(), w)

	// Send last login to db asynchronously
	go func() {
		err := ar.authService.UpdateLastLogin(user.Id)
		if err != nil {
			ar.logger.Error("Failed to update last login", gecho.Field("error", err), gecho.Field("userID", user.Id))
		}
	}()

	// clear password from user
	user.PasswordHash = ""

	gecho.Success(w,
		gecho.WithMessage("Login successful"),
		gecho.WithData(user),
		gecho.Send(),
	)
}
