package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/repositories"
	"go-backend/internal/services"
	"go-backend/pkg/util"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

func SendMessageHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Invalid request payload: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		messageRepo := repositories.SendMessageRepository(db)
		messageService := services.SendMessageService(messageRepo)

		if err := messageService.SendMessage(senderID, req.ReceiverID, req.Content); err != nil {
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
		response := map[string]any{
			"messages":   messages,
			"pagination": pagination,
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func GetConversationsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the request (this should be from a logged-in user, e.g., using a token)
		userID, err := GetUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		pagination := util.GetPagination(r)
		// Define a struct to hold the conversation data
		var conversations []models.ConversationInfo

		// Query the conversations with the most recent message timestamp and the latest message content (limited to 20 characters)
		err = db.Select(&conversations, `
			SELECT c.id AS conversation_id,
				CASE 
					WHEN c.user1_id = $1 THEN c.user2_id 
					ELSE c.user1_id 
				END AS user_id,
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
			WHERE c.user1_id = $1 OR c.user2_id = $1
			ORDER BY m.sent_at DESC NULLS LAST
			LIMIT $2 OFFSET $3;
		`, userID, pagination.Limit, pagination.Offset)

		if err != nil {
			log.Printf("Error fetching conversations: %v", err)
			http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
			return
		}

		totalCount := len(conversations)
		// Prepare the response with conversation ids, user ids, latest message timestamp, and content
		response := []map[string]any{}
		for _, conversation := range conversations {
			// Handle the case when last_message_at is NULL
			lastMessageAt := ""
			if conversation.LastMessageAt.Valid {
				lastMessageAt = conversation.LastMessageAt.Time.Format(time.RFC3339)
			}

			// Prepare the conversation item with the latest message content (limited to 20 characters)
			item := map[string]any{
				"conversation_id":      conversation.ConversationID,
				"user_id":              conversation.UserID,
				"last_message_at":      lastMessageAt,
				"last_message_content": conversation.LastMessageContent,
			}
			response = append(response, item)
		}
		pagination.TotalItems = totalCount
		if pagination.Limit > 0 {
			pagination.TotalPages = (totalCount + pagination.Limit - 1) / pagination.Limit
		}

		finalResponse := map[string]any{
			"conversations": response,
			"pagination":    pagination,
		}

		// Send the response in JSON format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(finalResponse); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
