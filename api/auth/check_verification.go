package auth

import (
	"mamabloemetjes_server/api/response"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

// HandleCheckVerification checks if a user's email is verified
func (ar *AuthRoutesManager) HandleCheckVerification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameters
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		ar.logger.Warn("Missing user_id parameter")
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.auth.missingUserId")))
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ar.logger.Warn("Invalid user_id format", gecho.Field("error", err), gecho.Field("user_id", userIDStr))
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.auth.invalidUserIdFormat")))
		return
	}

	// Get user by ID
	user, err := ar.authService.GetUserByID(userID)
	if err != nil {
		ar.logger.Error("Failed to get user by ID", gecho.Field("error", err), gecho.Field("user_id", userID))
		// Don't reveal if user exists or not for security reasons
		response.Send(gecho.Success(w, gecho.WithData(map[string]interface{}{
			"verified": false,
		})))
		return
	}

	if user == nil {
		ar.logger.Warn("User not found", gecho.Field("user_id", userID))
		response.Send(gecho.Success(w, gecho.WithData(map[string]interface{}{
			"verified": false,
		})))
		return
	}

	// Return verification status
	ar.logger.Info("Verification status checked", gecho.Field("user_id", userID), gecho.Field("verified", user.EmailVerified))
	response.Send(gecho.Success(w, gecho.WithData(map[string]interface{}{
		"verified": user.EmailVerified,
		"email":    user.Email,
	})))
}
