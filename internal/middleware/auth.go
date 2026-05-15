package middleware

import (
	"context"
	"net/http"
	"mini_chat/internal/models"
)

type contextKey string

const UserKey contextKey = "username"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		username, err := models.GetUserBySession(r.Context(), cookie.Value)
		if err != nil || username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), UserKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}