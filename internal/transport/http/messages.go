package http

import (
	"chat/internal/chat"
	"encoding/json"
	"net/http"
)

type MessagesHandler struct {
	repo *chat.Repository
}

func NewMessagesHandler(repo *chat.Repository) *MessagesHandler {
	return &MessagesHandler{
		repo: repo,
	}
}

func (h* MessagesHandler) GetMessages(
	w http.ResponseWriter, 
	r *http.Request,
) {
	msgs, err := h.repo.GetMessage()
	if err != nil {
		http.Error(w, "server error", 500)
		return
	}

	json.NewEncoder(w).Encode(msgs)
}