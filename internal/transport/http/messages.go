package http

import (
	"chat/internal/chat"
	"encoding/json"
	"net/http"
	"strconv"
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
	dialogIDStr := r.URL.Query().Get("dialog_id")
	
	if dialogIDStr == "" {
		http.Error(w, "dialog_id required", 400)
		return
	}

	dialogID, err := strconv.Atoi(dialogIDStr)
	if err != nil {
		http.Error(w, "invalid dialog_id", 400)
		return
	}

	msgs, err := h.repo.GetMessage(dialogID)
	if err != nil {
		http.Error(w, "server error", 500)
		return
	}

	json.NewEncoder(w).Encode(msgs)
}