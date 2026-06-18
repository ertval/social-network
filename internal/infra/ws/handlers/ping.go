package handlers

import (
	"encoding/json"

	ws "social-network/internal/infra/ws"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h *PingHandler) Handle(client *ws.Client, env ws.Envelope) {
	reply, _ := json.Marshal(ws.Envelope{
		Type:      ws.TypePong,
		RequestID: env.RequestID,
	})
	client.Send(reply)
}
