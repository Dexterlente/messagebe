package services

import (
	"go-backend/internal/repositories"
)

type MessageService struct {
	*repositories.MessageRepository
}

func SendMessageService(repo *repositories.MessageRepository) *MessageService {
	return &MessageService{MessageRepository: repo}
}
