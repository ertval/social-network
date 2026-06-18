package user

import (
	"time"
)

type User struct {
	CreatedAt time.Time
	AvatarURL *string
	Password  string
	Nickname  string
	Email     string
	Role      string
	ID        string
	FirstName string
	LastName  string
	Gender    string
	Age       int
}
