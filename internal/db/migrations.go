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

// Migrate Users Table
func migrateUsers(db *sqlx.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
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