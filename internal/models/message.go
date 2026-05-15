package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	DialogID  int64     `json:"dialog_id"`
	SenderID  int64     `json:"sender_id"`
	Text      string    `json:"text"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
