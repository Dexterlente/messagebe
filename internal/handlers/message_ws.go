package handlers

import (
	"go-backend/internal/models"
	"go-backend/internal/repositories"
	"go-backend/internal/services"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Adjust this logic as needed for your application
		return true
	},
}

// @Summary WebSocket chat
// @Description Connect to this endpoint via WebSocket to send/receive messages
// @Tags WS Messages
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {string} string "Unauthorized"
// @Router /ws-send [get]
// @Security ApiKeyAuth
// @Param Authorization header string true "	Bearer <token>"
func HandleWebSocket(db *sqlx.DB, clients map[int]*websocket.Conn, clientsMu *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		senderID, err := GetUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Register the connection
		clientsMu.Lock()
		clients[senderID] = conn
		clientsMu.Unlock()

		// Clean up on disconnect
		defer func() {
			clientsMu.Lock()
			delete(clients, senderID)
			clientsMu.Unlock()
		}()

		for {
			var msg models.SendMessageRequest
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("Read error: %v", err)
				break
			}

			// Save message to DB
			messageRepo := repositories.SendMessageRepository(db)
			messageService := services.SendMessageService(messageRepo)
			err := messageService.SendMessage(senderID, msg.ReceiverID, msg.Content)
			if err != nil {
				log.Printf("DB insert error: %v", err)
				continue
			}

			// Forward message to recipient if online
			clientsMu.RLock()
			receiverConn, online := clients[msg.ReceiverID]
			clientsMu.RUnlock()

			if online {
				// You can enrich this to include sender info or timestamps
				if err := receiverConn.WriteJSON(map[string]any{
					"sender_id": senderID,
					"content":   msg.Content,
					"timestamp": time.Now().Format(time.RFC3339),
				}); err != nil {
					log.Printf("Write error to user %d: %v", msg.ReceiverID, err)
				}
			}
		}
	}
}
