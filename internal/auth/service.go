package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(username, password string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return s.repo.Create(username, string(hash))
}

func (s *Service) Login(username, password string) error {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}

	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}