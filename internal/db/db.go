package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New() (*sqlx.DB, error) {
    db, err := connect()
    if err != nil {
        return nil, err
    }

    if err := runMigrations(db); err != nil {
        return nil, err
    }

    log.Println("Connected to the database successfully!")
    return db, nil
}

func connect() (*sqlx.DB, error) {
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is not set")
    }

    return sqlx.Connect("postgres", databaseURL)
}