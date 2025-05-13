package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(db *sqlx.DB) {
	// Swagger documentation
	http.Handle("/swagger/", httpSwagger.WrapHandler)
	http.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json") // Adjust path if needed
	})

	// user_handler.go
	http.HandleFunc("/users", GetUsers(db))
	http.HandleFunc("/user", CreateUser(db))
	http.HandleFunc("/change-password", ChangePasswordHandlerFunc(db))
	http.HandleFunc("/login", LoginHandlerFunc(db))
	http.HandleFunc("/validate-token", TokenValidationHandler)

	//message_handler.go
	http.HandleFunc("/messages", GetMessagesHandler(db))
	http.HandleFunc("/send-message", SendMessageHandler(db))
	http.HandleFunc("/conversations", GetConversationsHandler(db))
}
