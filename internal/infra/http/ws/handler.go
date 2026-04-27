package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: actual origin in production
		return true
	},
}

// Hnadler upgrades the connection and registers the client with the hub.
// getUserIDFromRequest is injected so this package has no auth dependency.
func Handler(
	hub *Hub,
	getUserIDFromRequest func(r *http.Request) (string, error),
	onMessage func(client *Client, msg []byte),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromRequest(r)
		if err != nil || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
			return
		}

		client := NewCient(userID, hub, conn)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump(onMessage)
	}
}
