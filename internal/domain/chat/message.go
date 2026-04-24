package chat

import "time"

type Message struct {
	ID              int
	ChatID          string
	SenderID        string
	Content         string
	CreatedAt       time.Time
	ClientMessageID *string
}
