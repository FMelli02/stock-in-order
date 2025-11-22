package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"stock-in-order/backend/internal/services"
)

// ForgotPassword handles POST /api/v1/users/forgot-password
func ForgotPassword(db *pgxpool.Pool, emailService *services.EmailService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Email string `json:"email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if input.Email == "" {
			http.Error(w, "email is required", http.StatusBadRequest)
			return
		}

		// Get user by email
		const qUser = `SELECT id, name FROM users WHERE email = $1`
		var userID int64
		var userName string
		err := db.QueryRow(context.Background(), qUser, input.Email).Scan(&userID, &userName)
		if err != nil {
			// Por seguridad, no revelamos si el email existe o no
			slog.Info("Password reset requested for non-existent email", "email", input.Email)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}

		// Generate random token (32 bytes = 64 hex chars)
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			slog.Error("Failed to generate random token", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		plainToken := hex.EncodeToString(tokenBytes)

		// Hash the token for storage
		hash := sha256.Sum256([]byte(plainToken))
		tokenHash := hex.EncodeToString(hash[:])

		// Token expires in 1 hour
		expiry := time.Now().Add(1 * time.Hour)

		// Delete any existing tokens for this user
		const qDelete = `DELETE FROM password_tokens WHERE user_id = $1`
		_, err = db.Exec(context.Background(), qDelete, userID)
		if err != nil {
			slog.Error("Failed to delete old tokens", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Store hashed token in database
		const qInsert = `
			INSERT INTO password_tokens (hash, user_id, expiry)
			VALUES ($1, $2, $3)`
		_, err = db.Exec(context.Background(), qInsert, tokenHash, userID, expiry)
		if err != nil {
			slog.Error("Failed to store password token", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Get frontend URL from environment or use default
		// TODO: Make this configurable via environment variable
		frontendURL := "http://localhost:5173" // Default for development
		resetLink := frontendURL + "/reset-password?token=" + plainToken

		// Send email with reset link
		emailData := map[string]interface{}{
			"user_name":    userName,
			"reset_link":   resetLink,
			"expiry_hours": 1,
		}

		err = emailService.SendPasswordResetEmail(input.Email, emailData)
		if err != nil {
			slog.Error("Failed to send password reset email",
				"email", input.Email,
				"error", err,
			)
			http.Error(w, "failed to send email", http.StatusInternalServerError)
			return
		}

		slog.Info("Password reset email sent",
			"email", input.Email,
			"userID", userID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		})
	}
}

// ResetPassword handles PUT /api/v1/users/reset-password
func ResetPassword(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if input.Token == "" || input.NewPassword == "" {
			http.Error(w, "token and new_password are required", http.StatusBadRequest)
			return
		}

		// Validate password strength
		if len(input.NewPassword) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "password must be at least 6 characters",
			})
			return
		}

		// Hash the provided token to look it up
		hash := sha256.Sum256([]byte(input.Token))
		tokenHash := hex.EncodeToString(hash[:])

		// Look up token in database
		const qToken = `
			SELECT user_id, expiry 
			FROM password_tokens 
			WHERE hash = $1`

		var userID int64
		var expiry time.Time
		err := db.QueryRow(context.Background(), qToken, tokenHash).Scan(&userID, &expiry)
		if err != nil {
			slog.Warn("Invalid or non-existent password reset token")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid or expired token",
			})
			return
		}

		// Check if token has expired
		if time.Now().After(expiry) {
			slog.Warn("Expired password reset token used", "userID", userID)
			// Delete expired token
			const qDeleteExpired = `DELETE FROM password_tokens WHERE hash = $1`
			_, _ = db.Exec(context.Background(), qDeleteExpired, tokenHash)

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "token has expired",
			})
			return
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Failed to hash password", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Update user's password
		const qUpdate = `UPDATE users SET password = $1 WHERE id = $2`
		_, err = db.Exec(context.Background(), qUpdate, string(hashedPassword), userID)
		if err != nil {
			slog.Error("Failed to update user password", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Delete the used token
		const qDelete = `DELETE FROM password_tokens WHERE hash = $1`
		_, err = db.Exec(context.Background(), qDelete, tokenHash)
		if err != nil {
			slog.Warn("Failed to delete used token", "error", err)
			// Non-critical error, continue
		}

		slog.Info("Password successfully reset", "userID", userID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "password successfully reset",
		})
	}
}
