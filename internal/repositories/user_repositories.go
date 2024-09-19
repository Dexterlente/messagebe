package repositories

import (
	"database/sql"
	"go-backend/internal/models"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(db *sqlx.DB) ([]models.User, error) {
    var users []models.User
    err := db.Select(&users, "SELECT id, first_name, last_name, email, username FROM users")
    return users, err
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