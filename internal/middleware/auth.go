package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"chat-app/internal/auth"
	"chat-app/internal/models"
)

type contextKey string

const userContextKey contextKey = "currentUser"

func CurrentUser(r *http.Request) (*models.User, bool) {
	v := r.Context().Value(userContextKey)
	if v == nil {
		return nil, false
	}
	u, ok := v.(*models.User)
	return u, ok
}

func loadUserBySession(db *sql.DB, cookieName string, r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		return nil, err
	}

	var user models.User
	err = db.QueryRow(`
        SELECT u.id, u.username, u.email, u.created_at
        FROM sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.token_hash = $1 AND s.expires_at > $2
    `, auth.HashToken(cookie.Value), time.Now()).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func AuthMiddleware(db *sql.DB, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := loadUserBySession(db, cookieName, r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PageAuthMiddleware(db *sql.DB, cookieName, loginPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := loadUserBySession(db, cookieName, r)
			if err != nil {
				http.Redirect(w, r, loginPath, http.StatusFound)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
