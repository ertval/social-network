package chats

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/pkg/uuid"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func normalizeUserPair(a, b string) (string, string, error) {
	if a == "" || b == "" {
		return "", "", errors.New("user IDs cannot be empty")
	}
	if a == b {
		return "", "", errors.New("cannot create chat with self")
	}
	if a < b {
		return a, b, nil
	}
	return b, a, nil
}

// Think about  creating a chatID inside this function or in application layer.
func (r *Repo) GetOrCreateChat(ctx context.Context, userID1, userID2 string) (*chat.Chat, error) {
	lowID, highID, err := normalizeUserPair(userID1, userID2)
	if err != nil {
		return nil, err
	}

	chatID := uuid.NewProvider().NewUUID()

	// If chat already exists for this pair, INSERT is ignored because of UNIQUE constraint(user_low_id, user_high_id)
	_, err = r.DB.ExecContext(ctx, `
		INSERT OR IGNORE INTO direct_chats (id, user_low_id, user_high_id)
		VALUES (?, ?, ?)
	`, chatID, lowID, highID)
	if err != nil {
		return nil, err
	}

	var c chat.Chat
	var lastMessageID sql.NullInt64
	var lastMessageAt sql.NullTime

	err = r.DB.QueryRowContext(ctx, `
		SELECT id, user_low_id, user_high_id, created_at, updated_at, last_message_id, last_message_at
		FROM direct_chats
		WHERE user_low_id = ? AND user_high_id = ?
		LIMIT 1
	`, lowID, highID).Scan(
		&c.ID,
		&c.UserLowID,
		&c.UserHighID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&lastMessageID,
		&lastMessageAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This should never happen because we just inserted or found the chat
			return nil, errors.New("chat not found after create")
		}
		return nil, err
	}

	if lastMessageID.Valid {
		id := int(lastMessageID.Int64)
		c.LastMessageID = &id
	}
	if lastMessageAt.Valid {
		t := lastMessageAt.Time
		c.LastMessageAt = &t
	}

	return &c, nil

}
