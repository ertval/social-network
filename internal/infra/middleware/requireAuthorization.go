package middleware

import (
	"context"
	"net/http"

	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/pkg/helpers"
)

// type requireAuthMiddleware struct {
// 	sessionManager session.Manager
// }

// type RequireAuthMiddleware interface {
// 	Required(next http.HandlerFunc) http.HandlerFunc
// }

// func NewRequireAuthMiddleware(sessionManager session.Manager) AuthorizationInterface {
// 	return Authorization{
// 		sessionManager: sessionManager,
// 	}
// }

func (a authorization) Required(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken, refreshToken := GetTokensFromRequest(r)

		var (
			session *session.Session
			err     error
		)

		if sessionToken == "" && refreshToken != "" {
			session, err = a.sessionManager.GetSessionByRefreshToken(refreshToken)
		} else {
			session, err = a.sessionManager.GetSessionFromSessionTokens(sessionToken, refreshToken)
		}
		if err != nil {
			helpers.RespondWithJSON(
				w,
				http.StatusUnauthorized,
				nil,
				"Unauthorized: Invalid session")
			return
		}

		sessionExpired, refreshTokenExpired := CheckTokenExpiration(session)

		// If refresh token expired, reject (covers both "both expired" and "refresh expired" cases)
		if refreshTokenExpired {
			helpers.RespondWithError(w,
				http.StatusUnauthorized,
				"Unauthorized: Refresh token expired")
			return
		}

		// If access token expired but refresh is valid, create new session
		if sessionExpired {
			_ = a.sessionManager.DeleteSession(session.AccessToken)
			session, err = a.sessionManager.CreateSession(r.Context(), session.UserID)
			if err != nil || session == nil {
				helpers.RespondWithError(
					w,
					http.StatusUnauthorized,
					"Unauthorized: Could not refresh session")
				return
			}
			a.sessionManager.SetCookies(w, session)
		}

		// Both tokens valid, or just refreshed - get user and proceed
		user, err := a.sessionManager.GetUserFromSession(session.AccessToken)
		if err != nil {
			helpers.RespondWithError(
				w,
				http.StatusUnauthorized,
				"Unauthorized: User not found")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
