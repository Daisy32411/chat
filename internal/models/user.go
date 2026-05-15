package models

import (
	"context"
	"errors"
	"mini_chat/internal/db"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID		 int
	Username string
	Password string
}

func CreateUser(ctx context.Context, username, hashedPassword string) error {
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		username, hashedPassword)

	return err
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		"SELECT id, username, password FROM users WHERE username=$1",
		username).Scan(&u.ID, &u.Username, &u.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func CreateSession(ctx context.Context, sessionID, username string) error {
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO sessions (session_id, username, expires_at) VALUES ($1, $2, now() + interval '1 day')",
		sessionID, username)
	return err
}

func GetUserBySession(ctx context.Context, sessionID string) (string, error) {
	var username string
	err := db.Pool.QueryRow(ctx,
		"SELECT username FROM sessions WHERE session_id=$1 AND expires_at > now()",
		sessionID).Scan(&username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return username, err
}

func DeleteSession(ctx context.Context, sessionID string) error {
	_, err := db.Pool.Exec(ctx, "DELETE FROM sessions WHERE session_id=$1", sessionID)
	return err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetUserIDByUsername(ctx context.Context, username string) (int, error) {
	var id int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&id)
	return id, err
}