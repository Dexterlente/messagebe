package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// Run all migrations
func runMigrations(db *sqlx.DB) error {
    migrations := []func(*sqlx.DB) error{
        migrateUsers,
        migrateMessages,
        // migrateProducts,  // Example additional table migration
        // Add more migrations here
    }

    for _, migrate := range migrations {
        if err := migrate(db); err != nil {
            return err
        }
    }
    log.Println("Migrations completed successfully!")
    return nil
}

func tableExists(db *sqlx.DB, tableName string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = $1
    )`
    err := db.Get(&exists, query, tableName)
    if err != nil {
        return false, fmt.Errorf("failed to check if table exists: %v", err)
    }
    return exists, nil
}

// Migrate Users Table
func migrateUsers(db *sqlx.DB) error {
    // Check if the table already exists
    exists, err := tableExists(db, "users")
    if err != nil {
        return fmt.Errorf("failed to check if users table exists: %v", err)
    }

    // If the table exists, print a message and return
    if exists {
        log.Println("Users table already exists!")
        return nil
    }

    // Create the table if it doesn't exist
    _, err = db.Exec(`
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            first_name VARCHAR(100) NOT NULL,
            last_name VARCHAR(100) NOT NULL,
            email VARCHAR(100) NOT NULL UNIQUE,
            username VARCHAR(100) UNIQUE,
            password VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create users table: %v", err)
    }

    log.Println("Users table created successfully!")
    return nil
}


func migrateMessages(db *sqlx.DB) error {
    // Check if the table already exists
    exists, err := tableExists(db, "messages")
    if err != nil {
        return fmt.Errorf("failed to check if messages table exists: %v", err)
    }

    // If the table exists, print a message and return
    if exists {
        log.Println("messages table already exists!")
        return nil
    }

    // Create the table if it doesn't exist
    _, err = db.Exec(`
        CREATE TABLE  messages (
            id SERIAL PRIMARY KEY,
            sender_id INT REFERENCES users(id) ON DELETE CASCADE,
            receiver_id INT REFERENCES users(id) ON DELETE CASCADE,
            content TEXT NOT NULL,
            sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create messages table: %v", err)
    }

    log.Println("Messages table created successfully!")
    return nil
}
