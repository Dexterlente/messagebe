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

func (r *MessageRepository) ReceiverExists(receiverID int) (bool, error) {
	var exists bool
	err := r.DB.Get(&exists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, receiverID)
	return exists, err
}

func (r *MessageRepository) GetConversationID(user1, user2 int) (int, error) {
	var id int
	err := r.DB.Get(&id, `
		SELECT id FROM conversations 
		WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
	`, user1, user2)
	return id, err
}

func (r *MessageRepository) CreateConversation(user1, user2 int) (int, error) {
	var id int
	err := r.DB.Get(&id, `
		INSERT INTO conversations (user1_id, user2_id)
		VALUES ($1, $2)
		RETURNING id
	`, user1, user2)
	return id, err
}

func (r *MessageRepository) InsertMessage(convoID, senderID, receiverID int, content string) error {
	_, err := r.DB.Exec(`
		INSERT INTO messages (conversation_id, sender_id, receiver_id, content)
		VALUES ($1, $2, $3, $4)
	`, convoID, senderID, receiverID, content)
	return err
}
