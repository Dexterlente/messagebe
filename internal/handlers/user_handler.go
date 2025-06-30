package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-backend/internal/models"
	"go-backend/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var mySigningKey = []byte("your_secret_key")

// GetUserID extracts the user ID from the Authorization header.
func GetUserID(r *http.Request) (int, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return 0, errors.New("missing Authorization header")
	}

	token, err := ValidateToken(tokenString)
	if err != nil {
		return 0, errors.New("invalid token: " + err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["user_id"].(float64); ok {
			return int(uid), nil
		}
		return 0, errors.New("user_id not found in token")
	}

	return 0, errors.New("invalid token claims")
}

func GetUserById(db *sqlx.DB, userID int) (*models.User, error) {
	var user models.User
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
func GetUserByIDHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query param "id"
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		// Convert to int
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid id parameter", http.StatusBadRequest)
			return
		}

		// Call your DB function
		user, err := GetUserById(db, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		img := ""
		if user.ImageProfile.Valid {
			img = user.ImageProfile.String
		}

		response := models.UserDetailReponse{
			ID:           user.ID,
			Username:     user.UserName,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			ImageProfile: img,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetUsers handles GET /users
// @Summary Get all users
// @Description Returns a list of all users (JWT-protected)
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.UserList
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /users [get]
func GetUsers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			ErrorResponse(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

// CreateUser creates a new user.
// @Summary      Create a new user
// @Description  Accepts a JSON payload to create a new user and returns the new user ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      models.CreateUserPayload  true  "User data"
// @Success      201   {object}  models.CreateUserResponse  "Created user ID"
// @Router       /create-user [post]
func CreateUser(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
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

		JSONResponse(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"id":      id,
		})
	}
}

// ChangePasswordHandlerFunc handles password change requests.
// @Summary      Change user password
// @Description  Accepts a JSON payload to change the user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      models.ChangePasswordRequest  true  "User data"
// @Success      200   {object}  models.ChangePasswordResponse  "Password changed successfully"
// @Failure      400   {object}  models.ErrorResponse  "Bad Request"
// @Failure      401   {object}  models.ErrorResponse  "Unauthorized"
// @Failure      404   {object}  models.ErrorResponse  "User not found"
// @Failure      500   {object}  models.ErrorResponse  "Internal Server Error"
// @Router       /change-password [post]
func ChangePasswordHandlerFunc(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
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

// LoginHandlerFunc handles user login requests.
// @Summary      User login
// @Description  Accepts a JSON payload to log in the user and returns a JWT token
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user  body      models.LoginRequest  true  "User credentials"
// @Success      200   {object}  models.LoginResponse  "JWT token and user ID"
// @Failure      400   {object}  models.ErrorResponse  "Bad Request"
// @Failure      401   {object}  models.ErrorResponse  "Unauthorized"
// @Failure      500   {object}  models.ErrorResponse  "Internal Server Error"
// @Router       /login [post]
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
		claims["user_id"] = user.ID
		claims["username"] = user.UserName
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		t, err := token.SignedString(mySigningKey)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Could not create token")
			return
		}

		JSONResponse(w, http.StatusOK, map[string]string{"token": t, "id": strconv.Itoa(user.ID)})
	}
}

// ValidateToken validates the JWT token and checks its expiration.
// @Summary      Validate JWT token
// @Description  Validates the JWT token and checks its expiration
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        token  header      string  true  "JWT token"
// @Success      200    {object}   models.TokenValidationResponse  "Token is valid"
// @Failure      401    {object}   models.ErrorResponse  "Unauthorized"
// @Failure      400    {object}   models.ErrorResponse  "Bad Request"
// @Failure      500    {object}   models.ErrorResponse  "Internal Server Error"
// @Router       /validate-token [get]
// @Security ApiKeyAuth
// @Security Bearer
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

// SearchUsersHandler handles searching for users by username.
// @Summary      Search users by username
// @Description  Searches for users by username
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        username  query      string  true  "Partial username"
// @Success      200       {array}    models.UserSearch  "List of users"
// @Failure      400       {object}   models.ErrorResponse  "Bad Request"
// @Failure	  500       {object}   models.ErrorResponse  "Internal Server Error"
// @Router       /search-users [get]
// @Security ApiKeyAuth
// @Security Bearer
func SearchUsersHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		partial := r.URL.Query().Get("username")
		if partial == "" {
			http.Error(w, "username param is required", http.StatusBadRequest)
			return
		}

		users, err := services.SearchUsersByUsername(db, partial)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, "failed to encode json", http.StatusInternalServerError)
		}
	}
}
