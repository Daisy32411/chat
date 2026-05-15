package service

import (
	"context"

	"chat-app/internal/models"
	"chat-app/internal/repository"
)

type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Search(ctx context.Context, query string) ([]models.User, error) {
	return s.repo.Search(ctx, query, 20)
}
