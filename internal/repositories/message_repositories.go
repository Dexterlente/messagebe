package repositories

import (
	"fmt"
	"go-backend/internal/models"

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

func GetMessagesRepository(db *sqlx.DB, senderID, receiverID, limit, offset int) ([]models.Message, int, int, error) {
	// Get conversation ID
	var conversationID int
	err := db.Get(&conversationID, `
		SELECT id 
		FROM conversations 
		WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
	`, senderID, receiverID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("conversation not found: %w", err)
	}

	// Get total message count
	var totalCount int
	err = db.Get(&totalCount, `
		SELECT COUNT(*) 
		FROM messages 
		WHERE conversation_id = $1
	`, conversationID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("count query failed: %w", err)
	}

	// Get messages
	var messages []models.Message
	err = db.Select(&messages, `
		SELECT id, conversation_id, sender_id, receiver_id, content, sent_at
		FROM messages 
		WHERE conversation_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`, conversationID, limit, offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("select query failed: %w", err)
	}

	return messages, totalCount, conversationID, nil
}

func GetConversationsRepository(db *sqlx.DB, userID, limit, offset int) ([]models.ConversationInfo, error) {
	var conversations []models.ConversationInfo
	err := db.Select(&conversations, `
		SELECT c.id AS conversation_id,
			CASE 
				WHEN c.user1_id = $1 THEN c.user2_id 
				ELSE c.user1_id 
			END AS user_id,
			u.image_profile,
			m.sent_at AS last_message_at,
			LEFT(m.content, 20) AS last_message_content
		FROM conversations c
		LEFT JOIN LATERAL (
			SELECT content, sent_at
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY sent_at DESC
			LIMIT 1
		) m ON true
		JOIN users u ON u.id = CASE 
			WHEN c.user1_id = $1 THEN c.user2_id 
			ELSE c.user1_id 
		END
		WHERE c.user1_id = $1 OR c.user2_id = $1
		ORDER BY m.sent_at DESC NULLS LAST
		LIMIT $2 OFFSET $3;
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return conversations, nil
}
