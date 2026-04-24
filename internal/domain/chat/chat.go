package chat

import "time"

type Chat struct {
	ID            string
	UserLowID     string
	UserHighID    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastMessageID *int
	LastMessageAt *time.Time
}
