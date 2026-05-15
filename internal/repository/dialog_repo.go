package repository

import (
	"context"
	"database/sql"

	"chat-app/internal/models"
)

type DialogListItem struct {
	DialogID    int64           `json:"dialog_id"`
	OtherUser   models.User     `json:"other_user"`
	LastMessage *models.Message `json:"last_message,omitempty"`
}

type DialogRepo struct {
	db *sql.DB
}

func NewDialogRepo(db *sql.DB) *DialogRepo { return &DialogRepo{db: db} }

func (r *DialogRepo) GetOrCreateDirectDialog(ctx context.Context, userA, userB int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
        SELECT d.id
        FROM dialogs d
        JOIN dialog_users du1 ON du1.dialog_id = d.id AND du1.user_id = $1
        JOIN dialog_users du2 ON du2.dialog_id = d.id AND du2.user_id = $2
        LIMIT 1
    `, userA, userB).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = r.db.QueryRowContext(ctx, `INSERT INTO dialogs DEFAULT VALUES RETURNING id`).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = r.db.ExecContext(ctx, `
        INSERT INTO dialog_users (dialog_id, user_id)
        VALUES ($1, $2), ($1, $3)
    `, id, userA, userB)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *DialogRepo) ListForUser(ctx context.Context, userID int64) ([]DialogListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT
            d.id,
            o.id,
            o.username,
            o.email,
            o.created_at,
            lm.id,
            lm.sender_id,
            lm.text,
            lm.is_read,
            lm.created_at
        FROM dialogs d
        JOIN dialog_users du ON du.dialog_id = d.id AND du.user_id = $1
        JOIN dialog_users du2 ON du2.dialog_id = d.id AND du2.user_id <> $1
        JOIN users o ON o.id = du2.user_id
        LEFT JOIN LATERAL (
            SELECT id, sender_id, text, is_read, created_at
            FROM messages m
            WHERE m.dialog_id = d.id
            ORDER BY created_at DESC
            LIMIT 1
        ) lm ON true
        ORDER BY COALESCE(lm.created_at, d.created_at) DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DialogListItem, 0)
	for rows.Next() {
		var item DialogListItem
		var msgID sql.NullInt64
		var senderID sql.NullInt64
		var text sql.NullString
		var isRead sql.NullBool
		var msgCreated sql.NullTime
		if err := rows.Scan(
			&item.DialogID,
			&item.OtherUser.ID,
			&item.OtherUser.Username,
			&item.OtherUser.Email,
			&item.OtherUser.CreatedAt,
			&msgID, &senderID, &text, &isRead, &msgCreated,
		); err != nil {
			return nil, err
		}
		if msgID.Valid {
			item.LastMessage = &models.Message{
				ID:        msgID.Int64,
				DialogID:  item.DialogID,
				SenderID:  senderID.Int64,
				Text:      text.String,
				IsRead:    isRead.Bool,
				CreatedAt: msgCreated.Time,
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
