package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"chat/internal/dialogs"
	"chat/internal/middleware"
)

type DialogHandler struct {
	service *dialogs.Service
}

type createReq struct {
	Username string `json:"username"`
}

func NewDialogHandler(service *dialogs.Service) *DialogHandler {
	return &DialogHandler{service: service}
}

func (h *DialogHandler) GetDialogs(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.Get(username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []dialogs.Dialog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *DialogHandler) CreateDialog(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body createReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	target := strings.TrimSpace(body.Username)
	if target == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	if target == username {
		http.Error(w, "cannot create dialog with yourself", http.StatusBadRequest)
		return
	}

	id, err := h.service.Create(username, target)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"dialog_id": id,
	})
}