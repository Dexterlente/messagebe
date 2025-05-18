package handlers

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
)

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func RegisterRoutes(db *sqlx.DB) {
	// Swagger documentation
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.DocExpansion("list"),       // Collapse all sections
		httpSwagger.DeepLinking(true),          // Enable URL deep linking
		httpSwagger.PersistAuthorization(true), // Persist authorization header
		httpSwagger.UIConfig(map[string]string{
			"defaultModelsExpandDepth": "-1",
		}), // Disable default models section
	))

	http.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// user_handler.go
	http.HandleFunc("/users", withCORS(GetUsers(db)))
	http.HandleFunc("/create-user", CreateUser(db))
	http.HandleFunc("/change-password", ChangePasswordHandlerFunc(db))
	http.HandleFunc("/login", LoginHandlerFunc(db))
	http.HandleFunc("/validate-token", TokenValidationHandler)
	http.HandleFunc("/search-users", SearchUsersHandler(db))

	//message_handler.go
	http.HandleFunc("/messages", GetMessagesHandler(db))
	http.HandleFunc("/send-message", SendMessageHandler(db))
	http.HandleFunc("/conversations", GetConversationsHandler(db))

	//message_ws.go
	http.HandleFunc("/ws-convo", HandleWebSocket(db, make(map[int]*websocket.Conn), &sync.RWMutex{}))
}
