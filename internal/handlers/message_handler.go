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
		var msg models.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Check if both sender and receiver exist
		var senderExists, receiverExists bool
		err := db.Get(&senderExists, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", msg.SenderID)
		if err != nil || !senderExists {
			http.Error(w, "Sender not found", http.StatusNotFound)
			return
		}

		err = db.Get(&receiverExists, "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)", msg.ReceiverID)
		if err != nil || !receiverExists {
			http.Error(w, "Receiver not found", http.StatusNotFound)
			return
		}

		// Insert the message into the database
		_, err = db.Exec("INSERT INTO messages (sender_id, receiver_id, content) VALUES ($1, $2, $3)",
			msg.SenderID, msg.ReceiverID, msg.Content)
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

		response := map[string]interface{}{
			"messages":   messages,
			"pagination": pagination,
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
