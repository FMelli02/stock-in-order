package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/crypto"
)

// Integration representa una integración de un usuario con una plataforma externa
type Integration struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Platform       string    `json:"platform"` // 'mercadolibre', 'shopify', etc.
	ExternalUserID *string   `json:"external_user_id,omitempty"`
	AccessToken    string    `json:"access_token,omitempty"`  // Solo en memoria, no en JSON por seguridad
	RefreshToken   string    `json:"refresh_token,omitempty"` // Solo en memoria, no en JSON por seguridad
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// IntegrationModel maneja las operaciones de base de datos para integraciones
type IntegrationModel struct {
	DB            *pgxpool.Pool
	EncryptionKey string
}

// Insert crea una nueva integración para un usuario
// Los tokens se encriptan antes de guardarlos en la base de datos
func (m *IntegrationModel) Insert(integration *Integration) error {
	ctx := context.Background()

	// Encriptar los tokens
	encryptedAccessToken, err := crypto.Encrypt(integration.AccessToken, m.EncryptionKey)
	if err != nil {
		return err
	}

	var encryptedRefreshToken []byte
	if integration.RefreshToken != "" {
		encryptedRefreshToken, err = crypto.Encrypt(integration.RefreshToken, m.EncryptionKey)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO integrations (user_id, platform, external_user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err = m.DB.QueryRow(ctx, query,
		integration.UserID,
		integration.Platform,
		integration.ExternalUserID,
		encryptedAccessToken,
		encryptedRefreshToken,
		integration.ExpiresAt,
	).Scan(&integration.ID, &integration.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetByUserAndPlatform obtiene la integración de un usuario para una plataforma específica
// Los tokens se desencriptan después de leerlos de la base de datos
func (m *IntegrationModel) GetByUserAndPlatform(userID int64, platform string) (*Integration, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, platform, external_user_id, access_token, refresh_token, expires_at, created_at
		FROM integrations
		WHERE user_id = $1 AND platform = $2`

	var integration Integration
	var encryptedAccessToken []byte
	var encryptedRefreshToken []byte

	err := m.DB.QueryRow(ctx, query, userID, platform).Scan(
		&integration.ID,
		&integration.UserID,
		&integration.Platform,
		&integration.ExternalUserID,
		&encryptedAccessToken,
		&encryptedRefreshToken,
		&integration.ExpiresAt,
		&integration.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Desencriptar los tokens
	accessToken, err := crypto.Decrypt(encryptedAccessToken, m.EncryptionKey)
	if err != nil {
		return nil, err
	}
	integration.AccessToken = accessToken

	if len(encryptedRefreshToken) > 0 {
		refreshToken, err := crypto.Decrypt(encryptedRefreshToken, m.EncryptionKey)
		if err != nil {
			return nil, err
		}
		integration.RefreshToken = refreshToken
	}

	return &integration, nil
}

// Update actualiza una integración existente
// Los tokens se encriptan antes de guardarlos
func (m *IntegrationModel) Update(integration *Integration) error {
	ctx := context.Background()

	// Encriptar los tokens
	encryptedAccessToken, err := crypto.Encrypt(integration.AccessToken, m.EncryptionKey)
	if err != nil {
		return err
	}

	var encryptedRefreshToken []byte
	if integration.RefreshToken != "" {
		encryptedRefreshToken, err = crypto.Encrypt(integration.RefreshToken, m.EncryptionKey)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE integrations
		SET external_user_id = $1,
		    access_token = $2,
		    refresh_token = $3,
		    expires_at = $4
		WHERE id = $5 AND user_id = $6`

	result, err := m.DB.Exec(ctx, query,
		integration.ExternalUserID,
		encryptedAccessToken,
		encryptedRefreshToken,
		integration.ExpiresAt,
		integration.ID,
		integration.UserID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete elimina una integración
func (m *IntegrationModel) Delete(userID int64, platform string) error {
	ctx := context.Background()

	query := `DELETE FROM integrations WHERE user_id = $1 AND platform = $2`

	result, err := m.DB.Exec(ctx, query, userID, platform)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetAllForUser obtiene todas las integraciones de un usuario
// NOTA: Los tokens NO se desencriptan en este método por seguridad
// Solo se retorna información básica de la integración
func (m *IntegrationModel) GetAllForUser(userID int64) ([]Integration, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, platform, external_user_id, expires_at, created_at
		FROM integrations
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []Integration
	for rows.Next() {
		var integration Integration
		err := rows.Scan(
			&integration.ID,
			&integration.UserID,
			&integration.Platform,
			&integration.ExternalUserID,
			&integration.ExpiresAt,
			&integration.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return integrations, nil
}

// IsTokenExpired verifica si el token de acceso ha expirado
func (i *Integration) IsTokenExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// UpsertByUserAndPlatform inserta o actualiza una integración
// Si ya existe una integración para ese usuario y plataforma, la actualiza
func (m *IntegrationModel) UpsertByUserAndPlatform(integration *Integration) error {
	existing, err := m.GetByUserAndPlatform(integration.UserID, integration.Platform)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// No existe, insertar
			return m.Insert(integration)
		}
		return err
	}

	// Ya existe, actualizar
	integration.ID = existing.ID
	return m.Update(integration)
}
