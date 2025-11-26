package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"stock-in-order/backend/internal/middleware"
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

		// El registro público SIEMPRE crea admins (cada registro es una organización nueva)
		// Los vendedores/repositores solo pueden ser creados por un admin vía CreateUserByAdmin
		role := "admin"
		if in.Role != "" && in.Role != "admin" {
			// Si se especifica un rol diferente, ignorarlo y forzar admin
			slog.Warn("Public registration attempted with non-admin role, forcing admin", "requested_role", in.Role)
		}

		user := &models.User{
			Name:           in.Name,
			Email:          in.Email,
			PasswordHash:   hash,
			Role:           role,
			OrganizationID: 0, // Se auto-asignará en Insert para admins
		}

		if err := store.Insert(user); err != nil {
			if err == models.ErrDuplicateEmail {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "email already exists"})
				return
			}
			slog.Error("Error creating user", "error", err.Error())
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
			"user_id":         user.ID,
			"email":           user.Email,          // Incluir el email para auditoría
			"role":            user.Role,           // Incluir el rol del usuario en el token
			"organization_id": user.OrganizationID, // Incluir organization_id para filtrado de datos
			"exp":             jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			"iat":             jwt.NewNumericDate(time.Now()),
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

		// Obtener organization_id del admin que está creando el usuario
		adminOrgID, ok := r.Context().Value("organization_id").(int64)
		if !ok {
			// Fallback: usar el user_id del admin
			adminUserID, _ := r.Context().Value("user_id").(int64)
			adminOrgID = adminUserID
			slog.Warn("organization_id not in context, using user_id as fallback", "user_id", adminUserID)
		}

		slog.Info("Creating user", "admin_org_id", adminOrgID, "new_user_role", in.Role, "new_user_email", in.Email)

		user := &models.User{
			Name:           in.Name,
			Email:          in.Email,
			PasswordHash:   hash,
			Role:           in.Role,    // Explicit role from admin
			OrganizationID: adminOrgID, // Heredar organization_id del admin creador
		}

		um := &models.UserModel{DB: db}
		if err := um.Insert(user); err != nil {
			slog.Error("Failed to create user", "error", err, "email", in.Email, "role", in.Role, "organization_id", adminOrgID)
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

// GetCurrentUser retorna la información del usuario autenticado
func GetCurrentUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener user_id del contexto (inyectado por JWTMiddleware)
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener usuario de la BD
		um := &models.UserModel{DB: db}
		user, err := um.GetByID(userID)
		if err != nil {
			if err == models.ErrNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "could not fetch user", http.StatusInternalServerError)
			return
		}

		// Responder con datos del usuario (sin password hash)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	}
}

// GetAllUsers retorna la lista de usuarios de la organización (solo para admin)
func GetAllUsers(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener organization_id del admin autenticado
		adminOrgID, ok := middleware.OrganizationIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		um := &models.UserModel{DB: db}
		users, err := um.GetAllByOrganization(adminOrgID)
		if err != nil {
			slog.Error("Failed to fetch users", "error", err, "organization_id", adminOrgID)
			http.Error(w, "could not fetch users", http.StatusInternalServerError)
			return
		}

		// Formatear respuesta sin password hashes
		response := make([]map[string]any, len(users))
		for i, user := range users {
			response[i] = map[string]any{
				"id":         user.ID,
				"name":       user.Name,
				"email":      user.Email,
				"role":       user.Role,
				"created_at": user.CreatedAt,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// DeleteUser elimina un usuario de la organización (solo para admin)
func DeleteUser(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener organization_id del admin autenticado
		adminOrgID, ok := r.Context().Value("organization_id").(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener user_id del admin (para evitar auto-eliminación)
		adminUserID, ok := r.Context().Value("user_id").(int64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Extraer ID del usuario a eliminar de la URL usando gorilla/mux
		vars := mux.Vars(r)
		userIDStr := vars["id"]
		if userIDStr == "" {
			http.Error(w, "user id required", http.StatusBadRequest)
			return
		}

		targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		// Prevenir auto-eliminación
		if targetUserID == adminUserID {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "cannot delete yourself"})
			return
		}

		um := &models.UserModel{DB: db}

		// Verificar que el usuario pertenece a la organización del admin
		targetUser, err := um.GetByID(targetUserID)
		if err != nil {
			if err == models.ErrNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "could not fetch user", http.StatusInternalServerError)
			return
		}

		if targetUser.OrganizationID != adminOrgID {
			http.Error(w, "user not found in your organization", http.StatusNotFound)
			return
		}

		// Eliminar usuario
		if err := um.Delete(targetUserID); err != nil {
			slog.Error("Failed to delete user", "error", err, "user_id", targetUserID)
			http.Error(w, "could not delete user", http.StatusInternalServerError)
			return
		}

		slog.Info("User deleted", "user_id", targetUserID, "admin_id", adminUserID, "organization_id", adminOrgID)

		w.WriteHeader(http.StatusNoContent)
	}
}
