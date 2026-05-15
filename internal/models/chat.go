package models

import (
	"context"
	"mini_chat/internal/db"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	FromUser  string    `json:"from_user"`
	ToUser    string    `json:"to_user"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Dialog struct {
	WithUser      string    `json:"with_user"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
}

// SaveMessage сохраняет сообщение в БД. toUsername - имя получателя.
func SaveMessage(ctx context.Context, fromUserID int, toUsername, text string) error {
	var toUserID int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", toUsername).Scan(&toUserID)
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx,
		"INSERT INTO messages (from_user_id, to_user_id, text) VALUES ($1, $2, $3)",
		fromUserID, toUserID, text)
	return err
}

// GetDialogs возвращает список пользователей, с которыми были сообщения у данного пользователя,
// вместе с последним сообщением.
func GetDialogs(ctx context.Context, userID int) ([]Dialog, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT 
			other_user.username,
			m.text,
			m.created_at
		FROM (
			SELECT 
				CASE 
					WHEN from_user_id = $1 THEN to_user_id 
					ELSE from_user_id 
				END AS other_id,
				MAX(created_at) as last_time
			FROM messages
			WHERE from_user_id = $1 OR to_user_id = $1
			GROUP BY other_id
		) latest
		JOIN LATERAL (
			SELECT text, created_at
			FROM messages
			WHERE (from_user_id = $1 AND to_user_id = latest.other_id)
			   OR (from_user_id = latest.other_id AND to_user_id = $1)
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		JOIN users other_user ON other_user.id = latest.other_id
		ORDER BY m.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dialogs []Dialog
	for rows.Next() {
		var d Dialog
		if err := rows.Scan(&d.WithUser, &d.LastMessage, &d.LastMessageAt); err != nil {
			return nil, err
		}
		dialogs = append(dialogs, d)
	}
	return dialogs, nil
}

// GetMessagesBetween возвращает историю сообщений между userID и otherUsername
func GetMessagesBetween(ctx context.Context, userID int, otherUsername string) ([]Message, error) {
	var otherID int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", otherUsername).Scan(&otherID)
	if err != nil {
		return nil, err
	}
	rows, err := db.Pool.Query(ctx, `
		SELECT m.id, u_from.username, u_to.username, m.text, m.created_at
		FROM messages m
		JOIN users u_from ON m.from_user_id = u_from.id
		JOIN users u_to ON m.to_user_id = u_to.id
		WHERE (from_user_id = $1 AND to_user_id = $2)
		   OR (from_user_id = $2 AND to_user_id = $1)
		ORDER BY created_at ASC
	`, userID, otherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.FromUser, &msg.ToUser, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// SearchUsers ищет пользователей по подстроке username, исключая себя
func SearchUsers(ctx context.Context, userID int, query string) ([]string, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT username FROM users
		WHERE id != $1 AND username ILIKE '%' || $2 || '%'
		LIMIT 10
	`, userID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var usernames []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		usernames = append(usernames, u)
	}
	return usernames, nil
}