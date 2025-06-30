package models

import (
	"database/sql"
	"time"
)

type CreateUserResponse struct {
	ID      int  `json:"id"`
	Success bool `json:"success"`
}

type UserList struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID           int            `db:"id" json:"id"`
	FirstName    string         `db:"first_name" json:"first_name"`
	LastName     string         `db:"last_name" json:"last_name"`
	Email        string         `db:"email" json:"email"`
	UserName     string         `db:"username" json:"username"`
	ImageProfile sql.NullString `db:"image_profile" json:"image_profile"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

type UserDetailReponse struct {
	ID           int    `db:"id" json:"id"`
	Username     string `db:"username" json:"username"`
	FirstName    string `db:"first_name" json:"first_name"`
	LastName     string `db:"last_name" json:"last_name"`
	ImageProfile string `db:"image_profile" json:"image_profile"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

type LoginResponse struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
}

type Claims struct {
	Exp      float64 `json:"exp"`
	UserID   int     `json:"user_id"`
	Username string  `json:"username"`
}

type TokenValidationResponse struct {
	Message string `json:"message"`
	Claims  Claims `json:"claims"`
}

type MessageResponseSucess struct {
	Message string `json:"message"`
}

type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
	Count    int       `json:"count"`
	Total    int       `json:"total"`
	Offset   int       `json:"offset"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type GetConversationsResponse struct {
	Conversations []ConversationInfo `json:"conversations"`
	Count         int                `json:"count"`
	Total         int                `json:"total"`
	Offset        int                `json:"offset"`
	Page          int                `json:"page"`
	Limit         int                `json:"limit"`
}
