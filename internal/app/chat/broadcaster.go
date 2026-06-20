package chat

import "social-network/internal/domain/chat"

type Broadcaster interface {
	SendToUser(userID, requestID string, msg *chat.Message)
	IsOnline(userID string) bool
}
