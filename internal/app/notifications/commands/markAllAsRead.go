package notificationcommands

import (
	"context"

	"github.com/arnald/forum/internal/domain/notification"
)

type MarkAllAsReadRequest struct {
	UserID string
}

type MarkAllAsReadHandler interface {
	Handle(ctx context.Context, req MarkAllAsReadRequest) error
}

type markAllAsReadHandler struct {
	repo notification.Repository
}

func NewMarkAllAsReadHandler(repo notification.Repository) MarkAllAsReadHandler {
	return &markAllAsReadHandler{
		repo: repo,
	}
}

func (h *markAllAsReadHandler) Handle(ctx context.Context, req MarkAllAsReadRequest) error {
	return h.repo.MarkAllAsRead(ctx, req.UserID)
}
