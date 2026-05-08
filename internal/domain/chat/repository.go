package chat

import "context"

type Repository interface {
	// Get or create chat between two users
	GetOrCreateChat(ctx context.Context, userID1, userID2 string) (*Chat, error)

	// Get chat by ID
	GetChat(ctx context.Context, chatID string) (*Chat, error)

	// Get all chats for a user, sorted by last_message_at
	GetChatsForUser(ctx context.Context, userID string) ([]*Chat, error)

	// Send message in a chat
	SendMessage(ctx context.Context, chatID, senderID, content string, clientMessageID string) (*Message, error)

	// Get last 10 messages for a chat
	GetMessagesForChat(ctx context.Context, chatID string, limit int) ([]*Message, error)

	// Get messages before cursor for pagination
	GetMessagesForChatBefore(ctx context.Context, chatID string, beforeMessageID int, limit int) ([]*Message, error)

	// Mark messages as read
	MarkAsRead(ctx context.Context, chatID, userID string, upToMessageID int) error

	// Get unread message count for a chat and user
	GetUnreadCount(ctx context.Context, chatID, userID string) (int, error)

	// Get all unread counts for user's chats
	GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error)
}
