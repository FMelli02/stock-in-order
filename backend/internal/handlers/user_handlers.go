package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"stock-in-order/backend/internal/models"
)

var validate = validator.New()

// RegisterUserInput DTO for user registration.
type RegisterUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=admin vendedor repositor"`
}

// CreateUserByAdminInput DTO for admin creating a user with explicit role.
type CreateUserByAdminInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin vendedor repositor"`
}

// LoginUserInput DTO for user login.
type LoginUserInput struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// userInserter defines the behavior needed to insert a user. This facilitates testing.
type userInserter interface {
	Insert(user *models.User) error
}

// registerUserHandler returns a handler using the provided store for persistence.
func registerUserHandler(store userInserter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in RegisterUserInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}

		user := &models.User{
			Name:         in.Name,
			Email:        in.Email,
			PasswordHash: hash,
			Role:         in.Role, // Use role from input (or will default to 'vendedor' in DB)
		}

		if err := store.Insert(user); err != nil {
			if err == models.ErrDuplicateEmail {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "email already exists"})
				return
			}
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Do not include password hash in response
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	}
}

// RegisterUser returns an http.HandlerFunc that registers a new user.
func RegisterUser(db *pgxpool.Pool) http.HandlerFunc {
	store := &models.UserModel{DB: db}
	return registerUserHandler(store)
}

// LoginUser authenticates a user and returns a JWT token.
func LoginUser(db *pgxpool.Pool, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in LoginUserInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}

		um := &models.UserModel{DB: db}
		user, err := um.GetByEmail(in.Email)
		if err != nil {
			// Do not reveal whether the email exists
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(in.Password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create JWT token
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"role":    user.Role, // Incluir el rol del usuario en el token
			"exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			"iat":     jwt.NewNumericDate(time.Now()),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "could not create token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": signed,
			"user": map[string]any{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}

// CreateUserByAdmin creates a new user with explicit role assignment (Admin only).
func CreateUserByAdmin(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in CreateUserByAdminInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validate.Struct(in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}

		user := &models.User{
			Name:         in.Name,
			Email:        in.Email,
			PasswordHash: hash,
			Role:         in.Role, // Explicit role from admin
		}

		um := &models.UserModel{DB: db}
		if err := um.Insert(user); err != nil {
			if err == models.ErrDuplicateEmail {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "email already exists"})
				return
			}
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	}
}
