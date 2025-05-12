package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(db *sqlx.DB) {
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
