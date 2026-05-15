package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/middleware"
	"chat-app/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
	cfg config.Config
	db  *sql.DB
}

func NewAuthHandler(svc *service.AuthService, cfg config.Config, db *sql.DB) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg, db: db}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/login.html")
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/register.html")
}

func (h *AuthHandler) ChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html")
}

type authRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad json")
		return
	}

	user, err := h.svc.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	token, _, err := h.svc.CreateSession(r.Context(), user.ID, 7*24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "cannot create session")
		return
	}
	h.setCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad json")
		return
	}

	login := req.Login
	if login == "" {
		login = req.Username
	}
	user, err := h.svc.Login(r.Context(), login, req.Password)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, _, err := h.svc.CreateSession(r.Context(), user.ID, 7*24*time.Hour)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "cannot create session")
		return
	}
	h.setCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(h.cfg.CookieName); err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *AuthHandler) setCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
