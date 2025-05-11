package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/pkg/util"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func SendMessageHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SendMessageRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var receiverExists bool
		err = db.Get(&receiverExists, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", req.ReceiverID)
		if err != nil || !receiverExists {
			http.Error(w, "Receiver not found", http.StatusNotFound)
			return
		}

		_, err = db.Exec(`
			INSERT INTO messages (sender_id, receiver_id, content) 
			VALUES ($1, $2, $3)`,
			senderID, req.ReceiverID, req.Content,
		)
		if err != nil {
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully"})
	}
}

func GetMessagesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		receiverID := r.URL.Query().Get("receiver_id")
		if receiverID == "" {
			http.Error(w, "receiver_id is required", http.StatusBadRequest)
			return
		}

		pagination := util.GetPagination(r)

		var totalCount int
		err = db.Get(&totalCount, `
			SELECT COUNT(*) FROM messages 
			WHERE receiver_id = $1 AND sender_id = $2
		`, receiverID, senderID)
		if err != nil {
			http.Error(w, "Failed to count messages", http.StatusInternalServerError)
			return
		}

		var messages []models.Message
		err = db.Select(&messages, `
			SELECT * FROM messages 
			WHERE receiver_id = $1 AND sender_id = $2
			ORDER BY sent_at DESC
			LIMIT $3 OFFSET $4
		`, receiverID, senderID, pagination.Limit, pagination.Offset)
		if err != nil {
			http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		pagination.TotalItems = totalCount
		if pagination.Limit > 0 {
			pagination.TotalPages = (totalCount + pagination.Limit - 1) / pagination.Limit
		}

		response := map[string]interface{}{
			"messages":   messages,
			"pagination": pagination,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
