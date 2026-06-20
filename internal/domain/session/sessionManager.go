package session

import (
	"context"

	"social-network/internal/domain/user"
)

type Manager interface {
	CreateSession(ctx context.Context, userID string) (*Session, error)
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	GetUserFromSession(sessionID string) (*user.User, error)
	GetSessionFromSessionTokens(sessionToken, refreshToken string) (*Session, error)
	GetSessionByRefreshToken(refreshToken string) (*Session, error)
	ValidateSession(sessionID string) error
	DeleteSessionWhenNewCreated(ctx context.Context, sessionID string, userID string) error
}
