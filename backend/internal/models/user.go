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
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PasswordHash   []byte    `json:"-"`
	Role           string    `json:"role"`            // admin, vendedor, repositor
	OrganizationID int64     `json:"organization_id"` // ID de la organización a la que pertenece
	CreatedAt      time.Time `json:"created_at"`
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
		INSERT INTO users (name, email, password_hash, role, organization_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	var id int64
	var createdAt time.Time

	// Si no se especifica rol, usar 'vendedor' por defecto
	role := user.Role
	if role == "" {
		role = "vendedor"
	}

	// Convertir organization_id=0 a NULL para la base de datos
	var orgID interface{}
	if user.OrganizationID == 0 {
		orgID = nil
	} else {
		orgID = user.OrganizationID
	}

	err = tx.QueryRow(ctx, qUser, user.Name, user.Email, user.PasswordHash, role, orgID).Scan(&id, &createdAt)
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

	// Si es admin y no tiene organization_id, asignarse a sí mismo
	if role == "admin" && user.OrganizationID == 0 {
		const qUpdateOrg = `UPDATE users SET organization_id = $1 WHERE id = $1`
		_, err = tx.Exec(ctx, qUpdateOrg, id)
		if err != nil {
			return err
		}
		user.OrganizationID = id
	}

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
		SELECT id, name, email, password_hash, role, organization_id, created_at
		FROM users
		WHERE email = $1`

	var u User
	err := m.DB.QueryRow(context.Background(), q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.OrganizationID, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID fetches a user by ID.
func (m *UserModel) GetByID(id int64) (*User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, organization_id, created_at
		FROM users
		WHERE id = $1`

	var u User
	err := m.DB.QueryRow(context.Background(), q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.OrganizationID, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
