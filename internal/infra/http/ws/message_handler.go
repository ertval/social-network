package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/arnald/forum/internal/domain/chat"
)

// Inbound message types (client -> server)
const (
	TypeChatSend    = "chat.send"
	TypeChatHistory = "chat.history"
	TypeMarkRead    = "chat.mark_read"
	TypePing        = "ping"
)

// Outbound message types (server -> client)
const (
	TypeChatMessage   = "chat.message"
	TypeHistoryResult = "chat.history_result"
	TypeError         = "error"
	TypePong          = "pong"
)

// Envelope is the wrapper for every WebSocker message.
type Envelope struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// Payloads for inbound messages
type SendPayload struct {
	ChatID          string `json:"chat_id"`
	Content         string `json:"content"`
	ClientMessageID string `json:"client_message_id,omitempty"`
}

type HistoryPayload struct {
	ChatID          string `json:"chat_id"`
	BeforeMessageID int    `json:"before_message_id,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type MarkReadPayload struct {
	ChatID        string `json:"chat_id"`
	UpToMessageID int    `json:"up_to_message_id"`
}

// Payloads for outbound messages
type MessagePayload struct {
	ID        int       `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

// MessageHandler returns the onMessage function injected into the WebSocket handler.
// chatRepo is chat.Repository interface.
func MessageHandler(hub *Hub, chatRepo chat.Repository) func(*Client, []byte) {
	return func(client *Client, raw []byte) {
		var env Envelope
		err := json.Unmarshal(raw, &env)
		if err != nil {
			sendError(client, "", "invalid message format")
			return
		}

		ctx := context.Background()

		switch env.Type {

		case TypePing:
			reply, _ := json.Marshal(Envelope{
				Type:      TypePong,
				RequestID: env.RequestID,
			})
			client.send <- reply

		case TypeChatSend:
			var p SendPayload
			err := json.Unmarshal(env.Payload, &p)
			if err != nil {
				sendError(client, env.RequestID, "invalid payload")
				return
			}
			if p.ChatID == "" || p.Content == "" {
				sendError(client, env.RequestID, "chat_id and content are required")
				return
			}

			msg, err := chatRepo.SendMessage(ctx, p.ChatID, client.UserID, p.Content, p.ClientMessageID)
			if err != nil {
				sendError(client, env.RequestID, "failed to send message")
				log.Printf("ws SendMessage error: %v", err)
				return
			}

			outPayload, _ := json.Marshal(MessagePayload{
				ID:        msg.ID,
				ChatID:    msg.ChatID,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				CreatedAt: msg.CreatedAt,
			})
			reply, _ := json.Marshal(Envelope{
				Type:      TypeChatMessage,
				RequestID: env.RequestID,
				Payload:   outPayload,
			})

			// Send back to sender as confirmation
			client.send <- reply

			chat, err := chatRepo.GetChat(ctx, p.ChatID)
			if err == nil {
				recipientID := chat.UserHighID
				if recipientID == client.UserID {
					recipientID = chat.UserLowID
				}
				hub.Send(recipientID, reply)
			}

		case TypeChatHistory:
			var p HistoryPayload
			err := json.Unmarshal(env.Payload, &p)
			if err != nil {
				sendError(client, env.RequestID, "invalid payload")
				return
			}

			limit := p.Limit
			if limit <= 0 || limit > 20 {
				limit = 10
			}

			var messages []*chat.Message
			var histErr error

			if p.BeforeMessageID > 0 {
				messages, histErr = chatRepo.GetMessagesForChatBefore(ctx, p.ChatID, p.BeforeMessageID, limit)
			} else {
				messages, histErr = chatRepo.GetMessagesForChat(ctx, p.ChatID, limit)
			}
			if histErr != nil {
				sendError(client, env.RequestID, "failed to load messages")
				return
			}

			outPayload, _ := json.Marshal(messages)
			reply, _ := json.Marshal(Envelope{
				Type:      TypeHistoryResult,
				RequestID: env.RequestID,
				Payload:   outPayload,
			})
			client.send <- reply

		case TypeMarkRead:
			var p MarkReadPayload
			err := json.Unmarshal(env.Payload, &p)
			if err != nil {
				sendError(client, env.RequestID, "invalid payload")
				return
			}
			err = chatRepo.MarkAsRead(ctx, p.ChatID, client.UserID, p.UpToMessageID)
			if err != nil {
				sendError(client, env.RequestID, "failed to mark as read")
				return
			}

		default:
			sendError(client, env.RequestID, "unknown message type")
		}

	}
}

func sendError(client *Client, requestID, message string) {
	payload, _ := json.Marshal(ErrorPayload{Message: message})
	reply, _ := json.Marshal(Envelope{
		Type:      TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.send <- reply
}
