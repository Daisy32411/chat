package service

import (
	"context"
	"errors"
	"strings"

	"chat-app/internal/models"
	"chat-app/internal/repository"
)

type ChatService struct {
	dialogs  *repository.DialogRepo
	messages *repository.MessageRepo
	users    *repository.UserRepo
}

func NewChatService(dialogs *repository.DialogRepo, messages *repository.MessageRepo, users *repository.UserRepo) *ChatService {
	return &ChatService{dialogs: dialogs, messages: messages, users: users}
}

func (s *ChatService) ListDialogs(ctx context.Context, userID int64) ([]repository.DialogListItem, error) {
	return s.dialogs.ListForUser(ctx, userID)
}

func (s *ChatService) UsersSearch(ctx context.Context, query string) ([]models.User, error) {
	return s.users.Search(ctx, query, 20)
}

func (s *ChatService) OpenDialog(ctx context.Context, meID, otherUserID int64) (int64, error) {
	if meID == otherUserID {
		return 0, errors.New("cannot chat with yourself")
	}
	if _, err := s.users.GetByID(ctx, otherUserID); err != nil {
		return 0, err
	}
	return s.dialogs.GetOrCreateDirectDialog(ctx, meID, otherUserID)
}

func (s *ChatService) ListMessages(ctx context.Context, dialogID int64) ([]models.Message, error) {
	return s.messages.ListByDialog(ctx, dialogID)
}

func (s *ChatService) SendMessage(ctx context.Context, dialogID, senderID int64, text string) (*models.Message, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("message is empty")
	}
	return s.messages.Create(ctx, dialogID, senderID, text)
}
