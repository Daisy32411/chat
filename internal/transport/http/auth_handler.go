package http

import (
	"encoding/json"
	"net/http"

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
	var req req
	json.NewDecoder(r.Body).Decode(&req)

	err := h.auth.Register(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(201)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req req
	json.NewDecoder(r.Body).Decode(&req)

	err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid", 401)
		return
	}

	token, _ := auth.GenerateToken(req.Username)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}