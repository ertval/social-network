package oauthlogin

import (
	"context"
	"fmt"
	oauthservice "github.com/arnald/forum/internal/app/oauth"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/pkg/helpers"
	oauthpkg "github.com/arnald/forum/internal/pkg/oAuth"
	"net/http"
)

type OAuthHandler struct {
	provider       oauthpkg.Provider
	config         *config.ServerConfig
	loginService   *oauthservice.OAuthService
	stateManager   *oauthpkg.StateManager
	sessionManager session.Manager
	cookieManager  *authcookies.Manager
	logger         logger.Logger
}

func NewOAuthHandler(
	provider oauthpkg.Provider,
	config *config.ServerConfig,
	loginService *oauthservice.OAuthService,
	stateManager *oauthpkg.StateManager,
	sessionManager session.Manager,
	logger logger.Logger,
	cookieManager *authcookies.Manager,
) *OAuthHandler {
	return &OAuthHandler{
		provider:       provider,
		config:         config,
		loginService:   loginService,
		stateManager:   stateManager,
		sessionManager: sessionManager,
		logger:         logger,
		cookieManager:  cookieManager,
	}
}

func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := h.stateManager.Generate()
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	authURL := h.provider.GetAuthURL(state)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		h.logger.PrintError(fmt.Errorf("%w: %s", ErrInParameters, errParam), nil)
		http.Error(
			w,
			"problem with oatuh, see logger",
			http.StatusInternalServerError,
		)
		return
	}

	if code == "" {
		h.logger.PrintError(ErrCodeMissing, nil)
		http.Error(
			w,
			"no code in callback",
			http.StatusInternalServerError,
		)
		return
	}

	err := h.stateManager.Verify(state)
	if err != nil {
		h.logger.PrintError(err, nil)
		http.Error(
			w,
			"problem with oauth STATE, SEE LOGGER",
			http.StatusInternalServerError,
		)
		return
	}

	user, err := h.loginService.Login(
		ctx,
		code,
		h.provider,
	)
	if err != nil {
		h.logger.PrintError(err, map[string]string{
			"action":   "oauth_login",
			"provider": h.provider.Name(),
		})
		http.Error(
			w,
			"error at github_login",
			http.StatusInternalServerError,
		)
		return
	}

	session, err := h.sessionManager.CreateSession(r.Context(), user.ID)
	if err != nil {
		helpers.RespondWithError(w,
			http.StatusInternalServerError,
			"error at creating session",
		)
		h.logger.PrintError(err, nil)
		return
	}

	h.cookieManager.SetCookies(w, session)

	// Determine the provider-specific frontend callback URL
	var frontendCallbackBase string
	switch h.provider.Name() {
	case "github":
		frontendCallbackBase = h.config.OAuth.GitHub.FrontendCallbackURL
	case "google":
		frontendCallbackBase = h.config.OAuth.Google.FrontendCallbackURL
	default:
		frontendCallbackBase = h.config.OAuth.FrontendCallbackURL
	}

	http.Redirect(w, r, frontendCallbackBase, http.StatusTemporaryRedirect)

	h.logger.PrintInfo(
		"User logged in via "+h.provider.Name(),
		map[string]string{
			"user_id":  user.ID,
			"username": user.Nickname,
			"provider": h.provider.Name(),
		})
}
