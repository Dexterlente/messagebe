package handlers

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter JWT token as: Bearer {token}

import (
	"net/http"

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
	http.Handle("/swagger/", httpSwagger.WrapHandler)
	http.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json") // Adjust path if needed
	})

	// user_handler.go
	http.HandleFunc("/users", withCORS(GetUsers(db)))
	http.HandleFunc("/create-user", CreateUser(db))
	http.HandleFunc("/change-password", ChangePasswordHandlerFunc(db))
	http.HandleFunc("/login", LoginHandlerFunc(db))
	http.HandleFunc("/validate-token", TokenValidationHandler)

	//message_handler.go
	http.HandleFunc("/messages", GetMessagesHandler(db))
	http.HandleFunc("/send-message", SendMessageHandler(db))
	http.HandleFunc("/conversations", GetConversationsHandler(db))
}
