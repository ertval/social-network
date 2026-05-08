package chat

import "time"

type ChatRead struct {
	ChatID            string
	UserID            string
	LastReadMessageID *int
	LastReadAt        *time.Time
	UnreadCount       int
	UpdatedAt         time.Time
}
