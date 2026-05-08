package middleware

import (
	"net/http"

	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
)

type authorization struct {
	sessionManager session.Manager
	cookieManager  *authcookies.Manager
}
type Authorization interface {
	Required(next http.HandlerFunc) http.HandlerFunc
	Optional(next http.HandlerFunc) http.HandlerFunc
}

func NewAuthorizationMiddleware(sessionManager session.Manager, cookieManager *authcookies.Manager) Authorization {
	return authorization{
		sessionManager: sessionManager,
		cookieManager:  cookieManager,
	}
}
