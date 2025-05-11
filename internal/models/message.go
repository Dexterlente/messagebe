package models

import (
	"time"
)

type Message struct {
	ID             int       `db:"id"`              // Pk of the message
	ConversationID int       `db:"conversation_id"` // Fk to the conversation
	SenderID       int       `db:"sender_id"`
	ReceiverID     int       `db:"receiver_id"`
	Content        string    `db:"content"`
	SentAt         time.Time `db:"sent_at"`
}

type SendMessageRequest struct {
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}
