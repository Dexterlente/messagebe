package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/repositories"
	"go-backend/internal/services"
	"go-backend/pkg/util"
	"log"
	"net/http"
	"strconv"
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
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		receiverIDStr := r.URL.Query().Get("receiver_id")
		if receiverIDStr == "" {
			http.Error(w, "receiver_id is required", http.StatusBadRequest)
			return
		}

		receiverID, err := strconv.Atoi(receiverIDStr)
		if err != nil {
			http.Error(w, "invalid receiver_id", http.StatusBadRequest)
			return
		}

		pagination := util.GetPagination(r)
		messages, totalCount, err := services.FetchMessagesService(db, senderID, receiverID, pagination.Limit, pagination.Offset)
		if err != nil {
			log.Printf("Error retrieving messages: %v", err)
			http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		pagination.TotalItems = totalCount
		if pagination.Limit > 0 {
			pagination.TotalPages = (totalCount + pagination.Limit - 1) / pagination.Limit
		}

		response := map[string]any{
			"messages":   messages,
			"pagination": pagination,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func GetConversationsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, err := GetUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pagination := util.GetPagination(r)

		conversations, err := services.GetConversationsService(db, userID, pagination.Limit, pagination.Offset)
		if err != nil {
			log.Printf("Error fetching conversations: %v", err)
			http.Error(w, "Error fetching conversations", http.StatusInternalServerError)
			return
		}

		totalCount := len(conversations)

		response := make([]map[string]any, 0, len(conversations))
		for _, conv := range conversations {
			lastMessageAt := ""
			if conv.LastMessageAt.Valid {
				lastMessageAt = conv.LastMessageAt.Time.Format(time.RFC3339)
			}
			item := map[string]any{
				"conversation_id":      conv.ConversationID,
				"user_id":              conv.UserID,
				"last_message_at":      lastMessageAt,
				"last_message_content": conv.LastMessageContent,
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(finalResponse); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
