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

// SendMessageHandler handles sending messages between users.
// @Summary Send a message
// @Description Send a message from one user to another
// @Tags Messages
// @Accept json
// @Produce json
// @Param message body models.SendMessageRequest true "Message content"
// @Success 201 {object} models.MessageResponseSucess "Message sent successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /send-message [post]
// @Security ApiKeyAuth
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

// GetMessagesHandler handles fetching messages between two users.
// @Summary Get messages
// @Description Get messages between two users
// @Tags Messages
// @Accept json
// @Produce json
// @Param receiver_id query int true "Receiver ID"
// @Success 200 {object} models.GetMessagesResponse "Messages retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /messages [get]
// @Security ApiKeyAuth
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param offset query int false "Offset for pagination"
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

// GetConversationsHandler handles fetching conversations for a user.
// @Summary Get conversations
// @Description Get conversations for a user
// @Tags Messages
// @Accept json
// @Produce json
// @Success 200 {object} models.GetConversationsResponse "Conversations retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /conversations [get]
// @Security ApiKeyAuth
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param offset query int false "Offset for pagination"
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
			if !conv.LastMessageAt.IsZero() {
				lastMessageAt = conv.LastMessageAt.Format(time.RFC3339)
			}
			img := ""
			if conv.ImageProfile.Valid {
				img = conv.ImageProfile.String
			}

			item := map[string]any{
				"conversation_id":      conv.ConversationID,
				"user_id":              conv.UserID,
				"last_message_at":      lastMessageAt,
				"last_message_content": conv.LastMessageContent,
				"image_profile":        img,
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
