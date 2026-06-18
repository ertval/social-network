package notificationcommands

import (
	"context"

	"social-network/internal/domain/notification"
)

type MarkAsReadRequest struct {
	NotificationID int
	UserID         string
}

type MarkAsReadHandler interface {
	Handle(ctx context.Context, req MarkAsReadRequest) error
}

type markAsReadHandler struct {
	repo notification.Repository
}

func NewMarkAsReadHandler(repo notification.Repository) MarkAsReadHandler {
	return &markAsReadHandler{
		repo: repo,
	}
}

func (h *markAsReadHandler) Handle(ctx context.Context, req MarkAsReadRequest) error {
	return h.repo.MarkAsRead(ctx, req.NotificationID, req.UserID)
}
