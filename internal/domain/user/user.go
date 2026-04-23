package user

import (
	"time"
)

type User struct {
	CreatedAt time.Time
	Password  string
	AvatarURL *string
	Nickname  string
	Email     string
	Role      string
	ID        string
	FirstName string
	LastName  string
	Age       int
	Gender    string
}
