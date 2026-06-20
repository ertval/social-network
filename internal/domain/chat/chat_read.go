package chat

import "time"

type ChatRead struct {
	UpdatedAt         time.Time
	LastReadMessageID *int
	LastReadAt        *time.Time
	ChatID            string
	UserID            string
	UnreadCount       int
}
