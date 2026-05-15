package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chat/internal/chat"
	"chat/internal/dialogs"
	"chat/internal/middleware"
)

type MessagesHandler struct {
	repo       *chat.Repository
	dialogRepo *dialogs.Repository
}

func NewMessagesHandler(repo *chat.Repository, dialogRepo *dialogs.Repository) *MessagesHandler {
	return &MessagesHandler{
		repo:       repo,
		dialogRepo: dialogRepo,
	}
}

func (h *MessagesHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dialogIDStr := r.URL.Query().Get("dialog_id")
	if dialogIDStr == "" {
		http.Error(w, "dialog_id required", http.StatusBadRequest)
		return
	}

	dialogID, err := strconv.Atoi(dialogIDStr)
	if err != nil {
		http.Error(w, "invalid dialog_id", http.StatusBadRequest)
		return
	}

	member, err := h.dialogRepo.IsMember(dialogID, username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	msgs, err := h.repo.GetMessages(dialogID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if msgs == nil {
		msgs = []chat.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}