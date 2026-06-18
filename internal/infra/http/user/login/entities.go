package userlogin

import (
	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/domain/session"
	"social-network/internal/infra/http/authcookies"
	"social-network/internal/infra/logger"
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
