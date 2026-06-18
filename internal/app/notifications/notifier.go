package notifications

import "social-network/internal/domain/notification"

type Notifier interface {
	RegisterClient(userID string) chan *notification.Notification
	UnregisterClient(userID string, ch chan *notification.Notification)
	BroadcastToUser(userID string, notification *notification.Notification)
}
