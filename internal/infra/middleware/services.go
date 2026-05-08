package middleware

import (
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
)

type Middleware struct {
	Authorization Authorization
}

func NewMiddleware(sessionManager session.Manager, cookieManager *authcookies.Manager) *Middleware {
	return &Middleware{
		Authorization: NewAuthorizationMiddleware(sessionManager, cookieManager),
	}
}
