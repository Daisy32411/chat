package repository

import (
	"context"
	"database/sql"

	"chat-app/internal/models"
)

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo { return &MessageRepo{db: db} }

func (r *MessageRepo) Create(ctx context.Context, dialogID, senderID int64, text string) (*models.Message, error) {
	var m models.Message
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO messages (dialog_id, sender_id, text)
        VALUES ($1, $2, $3)
        RETURNING id, dialog_id, sender_id, text, is_read, created_at
    `, dialogID, senderID, text).Scan(&m.ID, &m.DialogID, &m.SenderID, &m.Text, &m.IsRead, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MessageRepo) ListByDialog(ctx context.Context, dialogID int64) ([]models.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, dialog_id, sender_id, text, is_read, created_at
        FROM messages
        WHERE dialog_id = $1
        ORDER BY created_at ASC
    `, dialogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]models.Message, 0)
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.DialogID, &m.SenderID, &m.Text, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
