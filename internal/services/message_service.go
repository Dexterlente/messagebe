package services

import (
	"go-backend/internal/models"
	"go-backend/internal/repositories"

	"github.com/jmoiron/sqlx"
)

type MessageService struct {
	*repositories.MessageRepository
}

func SendMessageService(repo *repositories.MessageRepository) *MessageService {
	return &MessageService{MessageRepository: repo}
}

func FetchMessagesService(db *sqlx.DB, senderID, receiverID, limit, offset int) ([]models.Message, int, error) {
	messages, totalCount, _, err := repositories.GetMessagesRepository(db, senderID, receiverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return messages, totalCount, nil
}
