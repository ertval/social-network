package getchatusers

import (
	"net/http"
	"sort"
	"time"

	"github.com/arnald/forum/internal/app"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/infra/http/ws"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/pkg/helpers"
)

type Handler struct {
	services app.Services
	chatRepo chat.Repository
	hub      *ws.Hub
	logger   logger.Logger
}

func NewHandler(services app.Services, chatRepo chat.Repository, hub *ws.Hub, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		chatRepo: chatRepo,
		hub:      hub,
		logger:   logger,
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

	// 1. Get all users via existing query handler
	allUsers, err := h.services.UserServices.Queries.GetAllUsers.Handle(ctx)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// 2. Get my chats for last_message_at ordering
	myChats, err := h.chatRepo.GetChatsForUser(ctx, me.ID)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch chats")
		return
	}

	// Build lookup: otherUserID → last_message_at
	lastMsgAt := make(map[string]*time.Time)
	for _, c := range myChats {
		otherID := c.UserHighID
		if otherID == me.ID {
			otherID = c.UserLowID
		}
		lastMsgAt[otherID] = c.LastMessageAt
	}

	var withMsg []ChatUser
	var withoutMsg []ChatUser

	for _, u := range allUsers {
		if u.ID == me.ID {
			continue
		}
		cu := ChatUser{
			UserID:        u.ID,
			Nickname:      u.Nickname,
			IsOnline:      h.hub.IsOnline(u.ID),
			LastMessageAt: lastMsgAt[u.ID],
		}
		if cu.LastMessageAt != nil {
			withMsg = append(withMsg, cu)
		} else {
			withoutMsg = append(withoutMsg, cu)
		}
	}

	// withMsg: already sorted by last_message_at DESC from GetChatsForUser
	// withoutMsg: sort alphabetically by nickname
	sort.Slice(withoutMsg, func(i, j int) bool {
		return withoutMsg[i].Nickname < withoutMsg[j].Nickname
	})

	result := append(withMsg, withoutMsg...)
	helpers.RespondWithJSON(w, http.StatusOK, nil, result)
}
