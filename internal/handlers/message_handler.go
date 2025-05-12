package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/pkg/util"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func SendMessageHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON: %v", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Log the request payload
		log.Printf("Received message request: %v", req)

		// Get sender ID
		senderID, err := GetUserID(r)
		if err != nil {
			log.Printf("Failed to get sender ID: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check if receiver exists
		var receiverExists bool
		err = db.Get(&receiverExists, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", req.ReceiverID)
		if err != nil || !receiverExists {
			log.Printf("Receiver not found: %v", req.ReceiverID)
			http.Error(w, "Receiver not found", http.StatusNotFound)
			return
		}

		// Get or create conversation ID
		var conversationID int
		err = db.Get(&conversationID, `
            SELECT id FROM conversations 
            WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
        `, senderID, req.ReceiverID)

		if err != nil {
			// If conversation not found, create new conversation
			err = db.Get(&conversationID, `
                INSERT INTO conversations (user1_id, user2_id)
                VALUES ($1, $2)
                RETURNING id
            `, senderID, req.ReceiverID)
			if err != nil {
				log.Printf("Failed to create conversation: %v", err)
				http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
				return
			}
		}

		// Insert the message into the database
		_, err = db.Exec(`
			INSERT INTO messages (conversation_id, sender_id, receiver_id, content)
			VALUES ($1, $2, $3, $4)
		`, conversationID, senderID, req.ReceiverID, req.Content)

		if err != nil {
			log.Printf("Failed to send message: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully"})
	}
}

func GetMessagesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get sender ID from the request
		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Get receiver ID from the query parameter
		receiverID := r.URL.Query().Get("receiver_id")
		if receiverID == "" {
			http.Error(w, "receiver_id is required", http.StatusBadRequest)
			return
		}

		// Fetch pagination details
		pagination := util.GetPagination(r)

		// Check if a conversation exists between the sender and receiver
		var conversationID int
		err = db.Get(&conversationID, `
			SELECT id 
			FROM conversations 
			WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
		`, senderID, receiverID)

		if err != nil {
			log.Printf("Error fetching conversation: %v", err)
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		// Fetch total message count for pagination
		var totalCount int
		err = db.Get(&totalCount, `
			SELECT COUNT(*) 
			FROM messages 
			WHERE conversation_id = $1
		`, conversationID)
		if err != nil {
			log.Printf("Error counting messages: %v", err)
			http.Error(w, "Failed to count messages", http.StatusInternalServerError)
			return
		}

		// Retrieve the messages based on the conversation ID
		var messages []models.Message
		err = db.Select(&messages, `
			SELECT id, conversation_id, sender_id, receiver_id, content, sent_at
			FROM messages 
			WHERE conversation_id = $1
			ORDER BY sent_at DESC
			LIMIT $2 OFFSET $3
		`, conversationID, pagination.Limit, pagination.Offset)
		if err != nil {
			log.Printf("Error retrieving messages: %v", err)
			http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		// Update pagination info
		pagination.TotalItems = totalCount
		if pagination.Limit > 0 {
			pagination.TotalPages = (totalCount + pagination.Limit - 1) / pagination.Limit
		}

		// Prepare the response with messages and pagination details
		response := map[string]interface{}{
			"messages":   messages,
			"pagination": pagination,
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
