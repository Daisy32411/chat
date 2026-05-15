package middleware

import (
	"context"
	"net/http"
	"strings"

	"chat/internal/auth"
)

type contextKey string

const userKey contextKey = "user"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, claims.Username)
		next(w, r.WithContext(ctx))
	}
}

func UserFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userKey)
	username, ok := v.(string)
	return username, ok
}