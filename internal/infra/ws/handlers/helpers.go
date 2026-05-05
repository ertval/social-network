package handlers

import (
	"encoding/json"

	ws "github.com/arnald/forum/internal/infra/ws"
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
