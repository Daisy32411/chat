package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"chat/internal/auth"
)

type Handler struct {
	auth *auth.Service
}

func NewHandler(a *auth.Service) *Handler {
	return &Handler{auth: a}
}

type req struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.auth.Register(body.Username, body.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.auth.Login(body.Username, body.Password); err != nil {
		http.Error(w, "invalid", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(body.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	token = strings.TrimPrefix(token, "Bearer ")

	claims, err := auth.ParseToken(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": claims.Username,
	})
}