package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-backend/internal/models"
	"go-backend/internal/services"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var mySigningKey = []byte("your_secret_key") 

func RegisterRoutes(db *sqlx.DB) {
    http.HandleFunc("/users", GetUsers(db))
    http.HandleFunc("/user", CreateUser(db))
    http.HandleFunc("/change-password", ChangePasswordHandlerFunc(db))
    http.HandleFunc("/login", LoginHandlerFunc(db))
    http.HandleFunc("/validate-token", TokenValidationHandler)
}



func GetUsers(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            ErrorResponse(w, http.StatusUnauthorized, "Missing Authorization header")
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return mySigningKey, nil
        })

        if err != nil || !token.Valid {
            ErrorResponse(w, http.StatusUnauthorized, "Invalid token")
            return
        }

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

func LoginHandlerFunc(db *sqlx.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req models.LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            ErrorResponse(w, http.StatusBadRequest, err.Error())
            return
        }

        user, err := services.GetUserByUsername(db, req.UserName)
        if err != nil {
            ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
            return
        }

        err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
        if err != nil {
            ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
            return
        }

        token := jwt.New(jwt.SigningMethodHS256)
        claims := token.Claims.(jwt.MapClaims)
        claims["username"] = user.UserName
        claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

        t, err := token.SignedString(mySigningKey)
        if err != nil {
            ErrorResponse(w, http.StatusInternalServerError, "Could not create token")
            return
        }

        JSONResponse(w, http.StatusOK, map[string]string{"token": t})
    }
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return nil, fmt.Errorf("token expired")
			}
		}
		return token, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func TokenValidationHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
		return
	}

	token, err := ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Token is valid. Claims: %v", token.Claims)
}