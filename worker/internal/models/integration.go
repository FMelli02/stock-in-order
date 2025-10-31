package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/worker/internal/crypto"
)

var ErrNotFound = errors.New("record not found")

// Integration representa una integración de un usuario con una plataforma externa
type Integration struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Platform       string    `json:"platform"` // 'mercadolibre', 'shopify', etc.
	ExternalUserID *string   `json:"external_user_id,omitempty"`
	AccessToken    string    `json:"access_token,omitempty"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// IntegrationModel maneja las operaciones de base de datos para integraciones
type IntegrationModel struct {
	DB            *pgxpool.Pool
	EncryptionKey string
}

// GetByExternalUserID obtiene la integración usando el ID externo del usuario
func (m *IntegrationModel) GetByExternalUserID(externalUserID string, platform string) (*Integration, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, platform, external_user_id, expires_at, created_at
		FROM integrations
		WHERE external_user_id = $1 AND platform = $2`

	var integration Integration
	err := m.DB.QueryRow(ctx, query, externalUserID, platform).Scan(
		&integration.ID,
		&integration.UserID,
		&integration.Platform,
		&integration.ExternalUserID,
		&integration.ExpiresAt,
		&integration.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &integration, nil
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
