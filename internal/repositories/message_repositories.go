package repositories

import (
	"github.com/jmoiron/sqlx"
)

type MessageRepository struct {
	DB *sqlx.DB
}

func SendMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) SendMessage(senderID, receiverID int, content string) error {
	// Check if receiver exists
	var exists bool
	err := r.DB.Get(&exists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, receiverID)
	if err != nil || !exists {
		return err
	}

	// Get or create conversation
	var convoID int
	err = r.DB.Get(&convoID, `
		SELECT id FROM conversations 
		WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
	`, senderID, receiverID)

	if err != nil {
		// If no conversation exists, create one
		err = r.DB.Get(&convoID, `
			INSERT INTO conversations (user1_id, user2_id)
			VALUES ($1, $2)
			RETURNING id
		`, senderID, receiverID)
		if err != nil {
			return err
		}
	}

	// Insert message
	_, err = r.DB.Exec(`
		INSERT INTO messages (conversation_id, sender_id, receiver_id, content)
		VALUES ($1, $2, $3, $4)
	`, convoID, senderID, receiverID, content)

	return err
}
