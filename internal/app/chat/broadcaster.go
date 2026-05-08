package chat

import "github.com/arnald/forum/internal/domain/chat"

type Broadcaster interface {
	SendToUser(userID, requestID string, msg *chat.Message)
	IsOnline(userID string) bool
}
