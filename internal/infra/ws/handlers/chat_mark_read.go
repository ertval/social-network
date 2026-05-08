package handlers

import (
	"context"
	"encoding/json"

	chatcommands "github.com/arnald/forum/internal/app/chat/commands"
	"github.com/arnald/forum/internal/infra/logger"
	ws "github.com/arnald/forum/internal/infra/ws"
)

type ChatMarkReadHandler struct {
	markAsRead chatcommands.MarkAsReadHandler
	logger     logger.Logger
}

func NewChatMarkReadHandler(markAsReadHandler chatcommands.MarkAsReadHandler, logger logger.Logger) *ChatMarkReadHandler {
	return &ChatMarkReadHandler{
		markAsRead: markAsReadHandler,
		logger:     logger,
	}
}

func (h *ChatMarkReadHandler) Handle(client *ws.Client, env ws.Envelope) {
	var payload ws.MarkReadPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		h.logger.PrintError(err, nil)
		sendError(client, env.RequestID, "invalid payload")
		return
	}

	err := h.markAsRead.Handle(context.Background(), chatcommands.MarkAsReadRequest{
		ChatID:        payload.ChatID,
		UserID:        client.UserID,
		UpToMessageID: payload.UpToMessageID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		sendError(client, env.RequestID, "failed to mark as read")
	}
}
