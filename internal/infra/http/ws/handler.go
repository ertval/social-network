package ws

import (
	"net/http"

	"github.com/gorilla/websocket"

	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/ws"
	"social-network/internal/pkg/helpers"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: actual origin in production
		return true
	},
}

type Handler struct {
	hub    *ws.Hub
	router ws.WSRouter
	logger logger.Logger
}

func NewHandler(hub *ws.Hub, router ws.WSRouter, logger logger.Logger) *Handler {
	return &Handler{
		hub:    hub,
		router: router,
		logger: logger,
	}
}

func (h *Handler) UpgradeConnection(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.PrintError(err, nil)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}

	client := ws.NewClient(user.ID, h.hub, conn)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(h.router.Route)
}
