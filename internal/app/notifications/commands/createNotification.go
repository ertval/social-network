package notificationcommands

import (
	"context"

	"github.com/arnald/forum/internal/domain/notification"
)

type CreateNotificationRequest struct {
	Notification *notification.Notification
}

type CreateNotificationHandler interface {
	Handle(ctx context.Context, req CreateNotificationRequest) error
}

type createNotificationHandler struct {
	repo notification.Repository
}

func NewCreateNotificationHandler(repo notification.Repository) CreateNotificationHandler {
	return &createNotificationHandler{
		repo: repo,
	}
}

func (h *createNotificationHandler) Handle(ctx context.Context, req CreateNotificationRequest) error {
	return h.repo.Create(ctx, req.Notification)
}
