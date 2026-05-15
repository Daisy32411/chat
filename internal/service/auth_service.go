package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"chat-app/internal/auth"
	"chat-app/internal/models"
	"chat-app/internal/repository"
)

type AuthService struct {
	users    *repository.UserRepo
	sessions *repository.SessionRepo
}

func NewAuthService(users *repository.UserRepo, sessions *repository.SessionRepo) *AuthService {
	return &AuthService{users: users, sessions: sessions}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	if username == "" || email == "" || password == "" {
		return nil, errors.New("fill all fields")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	return s.users.Create(ctx, username, email, hash)
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*models.User, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return nil, errors.New("fill all fields")
	}

	user, passHash, err := s.users.GetByUsernameOrEmail(ctx, login)
	if err != nil {
		return nil, err
	}
	if err := auth.CheckPassword(passHash, password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (s *AuthService) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (string, time.Time, error) {
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(ttl)
	if err := s.sessions.Create(ctx, userID, tokenHash, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessions.DeleteByTokenHash(ctx, auth.HashToken(token))
}
