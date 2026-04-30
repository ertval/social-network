package userlogin

import (
	"github.com/arnald/forum/internal/app"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
	"github.com/arnald/forum/internal/infra/logger"
)

type Handler struct {
	UserServices   app.Services
	SessionManager session.Manager
	CookieManager  *authcookies.Manager
	Config         *config.ServerConfig
	Logger         logger.Logger
}

func NewHandler(config *config.ServerConfig, app app.Services, sm session.Manager, logger logger.Logger, cookieManager *authcookies.Manager) *Handler {
	return &Handler{
		UserServices:   app,
		SessionManager: sm,
		CookieManager:  cookieManager,
		Config:         config,
		Logger:         logger,
	}
}
