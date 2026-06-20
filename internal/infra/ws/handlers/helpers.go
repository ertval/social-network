package handlers

import (
	"encoding/json"

	ws "social-network/internal/infra/ws"
)

func sendError(client *ws.Client, requestID, message string) {
	payload, _ := json.Marshal(ws.ErrorPayload{Message: message})
	reply, _ := json.Marshal(ws.Envelope{
		Type:      ws.TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.Send(reply)
}
