package getchatusers

import (
	chatqueries "github.com/arnald/forum/internal/app/chat/queries"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/pkg/helpers"
	"net/http"
	"time"
)

type Handler struct {
	getChatUserQuery chatqueries.GetChatUsersHandler
	logger           logger.Logger
}

func NewHandler(getChatUserQuery chatqueries.GetChatUsersHandler, logger logger.Logger) *Handler {
	return &Handler{
		getChatUserQuery: getChatUserQuery,
		logger:           logger,
	}
}

type ChatUser struct {
	UserID        string     `json:"user_id"`
	Nickname      string     `json:"nickname"`
	IsOnline      bool       `json:"is_online"`
	LastMessageAt *time.Time `json:"last_message_at"`
}

func (h *Handler) GetChatUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	me := middleware.GetUserFromContext(r)
	if me == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()

	result, err := h.getChatUserQuery.Handle(ctx, chatqueries.GetChatUsersRequest{MeID: me.ID})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get Users chat")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, result)
}
