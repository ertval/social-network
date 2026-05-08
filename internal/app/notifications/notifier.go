package notifications

import "github.com/arnald/forum/internal/domain/notification"

type Notifier interface {
	RegisterClient(userID string) chan *notification.Notification
	UnregisterClient(userID string, ch chan *notification.Notification)
	BroadcastToUser(userID string, notification *notification.Notification)
}
