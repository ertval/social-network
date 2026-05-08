package logout

import (
	"net/http"

	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/pkg/helpers"
)

type Handler struct {
	sessionManager session.Manager
	cookieManager  *authcookies.Manager
	logger         logger.Logger
}

func NewHandler(sessionManager session.Manager, logger logger.Logger, cookieManager *authcookies.Manager) *Handler {
	return &Handler{
		sessionManager: sessionManager,
		cookieManager:  cookieManager,
		logger:         logger,
	}
}

// Logout deletes the user's session from the database.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionToken := h.cookieManager.DeleteCookies(r, w)
	if sessionToken == "" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "No session found")
		return
	}

	err := h.sessionManager.DeleteSession(sessionToken)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	h.logger.PrintInfo("User logged out successfully", map[string]string{
		"userId": user.ID,
		"name":   user.Nickname,
	})

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{
		"message": "Logged out successfully",
	})
}
