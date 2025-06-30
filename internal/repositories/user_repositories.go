package repositories

import (
	"database/sql"
	"fmt"
	"go-backend/internal/models"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(db *sqlx.DB) ([]models.UserResponse, error) {
	var users []models.UserResponse
	err := db.Select(&users, "SELECT id, first_name, last_name, email, username, image_profile FROM users")
	return users, err
}

func GetUserById(db *sqlx.DB, userID int) (*models.UserDetailReponse, error) {
	var user models.UserDetailReponse
	query := "SELECT id, username, first_name, last_name, image_profile FROM users WHERE id = $1"
	err := db.Get(&user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", userID)
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}
	return &user, nil
}

func CreateUser(db *sqlx.DB, user *models.User) (int, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int
	err = db.QueryRowx(`INSERT INTO users (first_name, last_name, email, password, username) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		user.FirstName, user.LastName, user.Email, string(hashedPassword), user.UserName).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func ChangePassword(db *sqlx.DB, req *models.ChangePasswordRequest) error {
	var user models.User
	err := db.Get(&user, "SELECT id, password FROM users WHERE id=$1", req.UserID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		log.Printf("Password mismatch error: %v", err)
		return bcrypt.ErrMismatchedHashAndPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing new password: %v", err)
		return err
	}

	_, err = db.Exec("UPDATE users SET password=$1, updated_at=$2 WHERE id=$3", hashedPassword, time.Now(), req.UserID)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		return err
	}

	log.Println("Password updated successfully")
	return nil
}

func GetUserByUsername(db *sqlx.DB, username string) (*models.User, error) {
	var user models.User
	query := "SELECT id, first_name, last_name, email, username, password, created_at, updated_at FROM users WHERE username=$1"
	err := db.Get(&user, query, username)
	if err != nil {
		log.Printf("Error fetching user with username %s: %v", username, err)
		return nil, err
	}
	return &user, nil
}

func SearchUsersByUsername(db *sqlx.DB, searchString string) ([]models.UserSearch, error) {
	var users []models.UserSearch
	searchTerm := "%" + strings.ToLower(searchString) + "%"
	query := `SELECT id, username, first_name, last_name FROM users WHERE username ILIKE $1`

	err := db.Select(&users, query, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}

	return users, nil
}
