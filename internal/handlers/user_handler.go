package handlers

import (
	"database/sql"
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/services"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)


func RegisterRoutes(db *sqlx.DB) {

    http.HandleFunc("/users", GetUsers(db))
    http.HandleFunc("/user", CreateUser(db))
    http.HandleFunc("/change-password", ChangePasswordHandlerFunc(db))
}



func GetUsers(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        users, err := services.GetUsers(db)
        if err != nil {
            ErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }
        JSONResponse(w, http.StatusOK, users)
    }
}

func CreateUser(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var user models.User
        if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
            ErrorResponse(w, http.StatusBadRequest, err.Error())
            return
        }

        id, err := services.CreateUser(db, &user)
        if err != nil {
            ErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        JSONResponse(w, http.StatusCreated, map[string]interface{}{"id": id})
    }
}

func ChangePasswordHandlerFunc(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req models.ChangePasswordRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        log.Printf("ChangePasswordRequest: %+v", req)

        err := services.ChangePassword(db, &req)
        if err != nil {
            switch err {
            case sql.ErrNoRows:
                http.Error(w, "User not found", http.StatusNotFound)
            case bcrypt.ErrMismatchedHashAndPassword:
                http.Error(w, "Old password is incorrect", http.StatusUnauthorized)
            default:
                http.Error(w, "Failed to change password", http.StatusInternalServerError)
            }
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
    }
}