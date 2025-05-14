package main

// @title Your API Title
// @version 1.0
// @description This is a Messaging app backend.

import (
	"go-backend/internal/db"
	"go-backend/internal/handlers"
	"log"
	"net/http"
	"os"

	_ "go-backend/docs"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	db, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	handlers.RegisterRoutes(db)

	log.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
