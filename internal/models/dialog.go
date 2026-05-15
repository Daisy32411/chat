package models

import "time"

type Dialog struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	OtherUser   User      `json:"other_user"`
	LastMessage *Message  `json:"last_message,omitempty"`
}
