package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"chat/internal/middleware"
	"chat/internal/users"
)

type UsersHandler struct {
	repo *users.Repository
}

func NewUsersHandler(repo *users.Repository) *UsersHandler {
	return &UsersHandler{repo: repo}
}

func (h *UsersHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	items, err := h.repo.Search(q, username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}