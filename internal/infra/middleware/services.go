package middleware

import (
	"social-network/internal/domain/session"
	"social-network/internal/infra/http/authcookies"
)

type Middleware struct {
	Authorization Authorization
}

func NewMiddleware(sessionManager session.Manager, cookieManager *authcookies.Manager) *Middleware {
	return &Middleware{
		Authorization: NewAuthorizationMiddleware(sessionManager, cookieManager),
	}
}
