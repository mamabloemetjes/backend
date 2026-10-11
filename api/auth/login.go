package auth

import (
	"mamabloemetjes_server/api/response"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.AuthRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate request body", gecho.Field("error", err))
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.auth.checkLoginInformation")))
		return
	}

	user, err := ar.authService.Login(body)
	if err != nil {
		ar.logger.Warn("Login failed", gecho.Field("error", err))
		response.Send(gecho.Unauthorized(w, gecho.WithMessage("error.auth.invalidCredentials")))
		return
	}

	if !user.EmailVerified {
		ar.logger.Warn("Email not verified", gecho.Field("userID", user.Id))
		response.Send(gecho.Forbidden(w, gecho.WithMessage("error.auth.verifyEmail"), gecho.WithData(user.Email)))
		return
	}

	accessToken, err := ar.authService.GenerateAccessToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate access token", gecho.Field("error", err))
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.auth.unableToCompleteLogin")))
		return
	}

	refreshToken, err := ar.authService.GenerateRefreshToken(user)
	if err != nil {
		ar.logger.Warn("Failed to generate refresh token", gecho.Field("error", err))
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.auth.unableToCompleteLogin")))
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
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.auth.login"),
		gecho.WithData(user),
	))
}
