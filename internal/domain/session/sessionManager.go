package session

import (
	"context"
	"net/http"

	"github.com/arnald/forum/internal/domain/user"
)

type Manager interface {
	CreateSession(ctx context.Context, userID string) (*Session, error)
	SetCookies(w http.ResponseWriter, session *Session)
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	DeleteCookies(r *http.Request, w http.ResponseWriter) (sessiontoken string)
	GetUserFromSession(sessionID string) (*user.User, error)
	GetSessionFromSessionTokens(sessionToken, refreshToken string) (*Session, error)
	ValidateSession(sessionID string) error
	DeleteSessionWhenNewCreated(ctx context.Context, sessionID string, userID string) error
}
