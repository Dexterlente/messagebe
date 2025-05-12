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
	// Check if receiver exists
	exists, err := s.Repo.ReceiverExists(receiverID)
	if err != nil || !exists {
		return err
	}

	// Get or create conversation
	convoID, err := s.Repo.GetConversationID(senderID, receiverID)
	if err != nil {
		convoID, err = s.Repo.CreateConversation(senderID, receiverID)
		if err != nil {
			return err
		}
	}

	// Insert message
	return s.Repo.InsertMessage(convoID, senderID, receiverID, content)
}
