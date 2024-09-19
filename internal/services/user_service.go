package services

import (
	"go-backend/internal/models"
	"go-backend/internal/repositories"

	"github.com/jmoiron/sqlx"
)

func GetUsers(db *sqlx.DB) ([]models.User, error) {
    return repositories.GetUsers(db)
}

func CreateUser(db *sqlx.DB, user *models.User) (int, error) {
    return repositories.CreateUser(db, user)
}

func ChangePassword(db *sqlx.DB, req *models.ChangePasswordRequest) error {
    return repositories.ChangePassword(db, req)
}

func GetUserByUsername(db *sqlx.DB, username string) (*models.User, error) {
	return repositories.GetUserByUsername(db, username)
}