package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a user record in the database.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	Role         string    `json:"role"` // admin, vendedor, repositor
	CreatedAt    time.Time `json:"created_at"`
}

// ErrDuplicateEmail is returned when inserting a user with an existing email.
var ErrDuplicateEmail = errors.New("duplicate email")

// UserModel wraps DB access for users.
type UserModel struct {
	DB *pgxpool.Pool
}

// Insert stores a new user and sets its ID and CreatedAt fields.
// También crea automáticamente una suscripción gratuita para el usuario.
func (m *UserModel) Insert(user *User) error {
	ctx := context.Background()

	// Iniciar transacción para crear usuario + suscripción atómicamente
	tx, err := m.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Crear usuario
	const qUser = `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	var id int64
	var createdAt time.Time

	// Si no se especifica rol, usar 'vendedor' por defecto
	role := user.Role
	if role == "" {
		role = "vendedor"
	}

	err = tx.QueryRow(ctx, qUser, user.Name, user.Email, user.PasswordHash, role).Scan(&id, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrDuplicateEmail
		}
		return err
	}

	user.ID = id
	user.CreatedAt = createdAt
	user.Role = role

	// 2. Crear suscripción gratuita para el nuevo usuario
	const qSubscription = `
		INSERT INTO subscriptions (user_id, plan_id, status, current_period_start)
		VALUES ($1, $2, $3, $4)
	`

	now := time.Now()
	_, err = tx.Exec(ctx, qSubscription, id, PlanFree, SubscriptionStatusActive, now)
	if err != nil {
		return err
	}

	// 3. Commit de la transacción
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetByEmail fetches a user by email.
func (m *UserModel) GetByEmail(email string) (*User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1`

	var u User
	err := m.DB.QueryRow(context.Background(), q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID fetches a user by ID.
func (m *UserModel) GetByID(id int64) (*User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE id = $1`

	var u User
	err := m.DB.QueryRow(context.Background(), q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
