package streamnotification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/infra/storage/notifications"
)

const tickerTime = 10

type Handler struct {
	service *notifications.NotificationService
}

func NewHandler(service *notifications.NotificationService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	userID := user.ID

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w,
			"Streaming unsupported",
			http.StatusInternalServerError,
		)
		return
	}

	notificationChan := h.service.RegisterClient(userID)
	defer h.service.UnregisterClient(userID, notificationChan)

	// connected
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err == nil {
		// unread count
		fmt.Fprintf(w, "event: unread_count\ndata: {\"type\":\"unread_count\",\"count\":%d}\n\n", count)
		flusher.Flush()
	}

	ticker := time.NewTicker(tickerTime * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// client disconected
			return
		case notification := <-notificationChan:
			data, err := json.Marshal(notification)
			if err != nil {
				continue
			}
			// notification
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			// heartbeat (comments are valid SSE — no change needed)
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
