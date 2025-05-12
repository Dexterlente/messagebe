package services

import (
	"go-backend/internal/repositories"
)

type MessageService struct {
	Repo *repositories.MessageRepository
}

func SendMessageService(repo *repositories.MessageRepository) *MessageService {
	return &MessageService{Repo: repo}
}

func (s *MessageService) SendMessage(senderID int, receiverID int, content string) error {
	return s.Repo.SendMessage(senderID, receiverID, content)
}
