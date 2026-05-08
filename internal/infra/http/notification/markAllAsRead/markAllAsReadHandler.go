package markallasread

import (
	"net/http"

	notificationcommands "github.com/arnald/forum/internal/app/notifications/commands"
	"github.com/arnald/forum/internal/infra/middleware"
)

type Handler struct {
	markAllAsRead notificationcommands.MarkAllAsReadHandler
}

func NewHandler(service notificationcommands.MarkAllAsReadHandler) *Handler {
	return &Handler{markAllAsRead: service}
}

func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	if user.ID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	userID := user.ID

	err := h.markAllAsRead.Handle(r.Context(), notificationcommands.MarkAllAsReadRequest{UserID: userID})
	if err != nil {
		http.Error(
			w,
			"failed to mark all as read",
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusOK)
}
