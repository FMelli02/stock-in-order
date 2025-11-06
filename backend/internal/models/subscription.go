package models

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// TIPOS Y CONSTANTES
// ============================================

// SubscriptionStatus representa el estado de una suscripción
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"   // Suscripción activa y pagada
	SubscriptionStatusInactive SubscriptionStatus = "inactive" // Sin pago, inactiva
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due" // Pago vencido
	SubscriptionStatusCanceled SubscriptionStatus = "canceled" // Cancelada
)

// PlanID representa los planes disponibles
type PlanID string

const (
	PlanFree       PlanID = "plan_free"       // Plan gratuito (limitado)
	PlanBasico     PlanID = "plan_basico"     // Plan básico ($)
	PlanPro        PlanID = "plan_pro"        // Plan profesional ($$)
	PlanEnterprise PlanID = "plan_enterprise" // Plan empresarial ($$$)
)

// PlanFeatures define las características de cada plan
type PlanFeatures struct {
	MaxProducts        int  // Máximo de productos
	MaxOrders          int  // Máximo de órdenes por mes
	MaxUsers           int  // Máximo de usuarios (multiusuario)
	AdvancedReports    bool // Reportes avanzados
	EmailNotifications bool // Notificaciones por email
	APIAccess          bool // Acceso a API
	PrioritySupport    bool // Soporte prioritario
	BatchTracking      bool // Trazabilidad de lotes
	ExpiryManagement   bool // Gestión de vencimientos
	MultiWarehouse     bool // Múltiples almacenes
	CustomIntegrations bool // Integraciones personalizadas
}

// GetPlanFeatures retorna las características de un plan
func GetPlanFeatures(planID PlanID) PlanFeatures {
	switch planID {
	case PlanFree:
		return PlanFeatures{
			MaxProducts:        50,
			MaxOrders:          20,
			MaxUsers:           1,
			AdvancedReports:    false,
			EmailNotifications: false,
			APIAccess:          false,
			PrioritySupport:    false,
			BatchTracking:      false,
			ExpiryManagement:   false,
			MultiWarehouse:     false,
			CustomIntegrations: false,
		}
	case PlanBasico:
		return PlanFeatures{
			MaxProducts:        200,
			MaxOrders:          100,
			MaxUsers:           3,
			AdvancedReports:    true,
			EmailNotifications: true,
			APIAccess:          false,
			PrioritySupport:    false,
			BatchTracking:      true,
			ExpiryManagement:   true,
			MultiWarehouse:     false,
			CustomIntegrations: false,
		}
	case PlanPro:
		return PlanFeatures{
			MaxProducts:        1000,
			MaxOrders:          500,
			MaxUsers:           10,
			AdvancedReports:    true,
			EmailNotifications: true,
			APIAccess:          true,
			PrioritySupport:    true,
			BatchTracking:      true,
			ExpiryManagement:   true,
			MultiWarehouse:     true,
			CustomIntegrations: false,
		}
	case PlanEnterprise:
		return PlanFeatures{
			MaxProducts:        -1, // Ilimitado
			MaxOrders:          -1, // Ilimitado
			MaxUsers:           -1, // Ilimitado
			AdvancedReports:    true,
			EmailNotifications: true,
			APIAccess:          true,
			PrioritySupport:    true,
			BatchTracking:      true,
			ExpiryManagement:   true,
			MultiWarehouse:     true,
			CustomIntegrations: true,
		}
	default:
		return GetPlanFeatures(PlanFree)
	}
}

// GetPlanPrice retorna el precio mensual de un plan en ARS
func GetPlanPrice(planID PlanID) float64 {
	switch planID {
	case PlanFree:
		return 0.0
	case PlanBasico:
		return 5000.0 // $5,000 ARS/mes
	case PlanPro:
		return 15000.0 // $15,000 ARS/mes
	case PlanEnterprise:
		return 40000.0 // $40,000 ARS/mes
	default:
		return 0.0
	}
}

// ============================================
// MODELOS
// ============================================

// Subscription representa una suscripción de usuario
type Subscription struct {
	ID                 int64              `json:"id"`
	UserID             int64              `json:"user_id"`
	PlanID             PlanID             `json:"plan_id"`
	MPSubscriptionID   sql.NullString     `json:"mp_subscription_id,omitempty"`
	MPCustomerID       sql.NullString     `json:"mp_customer_id,omitempty"`
	MPPreapprovalID    sql.NullString     `json:"mp_preapproval_id,omitempty"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time         `json:"current_period_end,omitempty"`
	TrialEnd           *time.Time         `json:"trial_end,omitempty"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	CancelReason       sql.NullString     `json:"cancel_reason,omitempty"`
	Metadata           map[string]any     `json:"metadata,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// PaymentHistory representa un registro de pago
type PaymentHistory struct {
	ID             int64     `json:"id"`
	SubscriptionID int64     `json:"subscription_id"`
	UserID         int64     `json:"user_id"`
	MPPaymentID    string    `json:"mp_payment_id"`
	MPStatus       string    `json:"mp_status"`
	MPStatusDetail string    `json:"mp_status_detail,omitempty"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Description    string    `json:"description"`
	PlanID         PlanID    `json:"plan_id"`
	PaymentDate    time.Time `json:"payment_date"`
	CreatedAt      time.Time `json:"created_at"`
}

// ============================================
// REPOSITORIO
// ============================================

// SubscriptionModel maneja operaciones de suscripciones
type SubscriptionModel struct {
	DB *pgxpool.Pool
}

// Create crea una nueva suscripción para un usuario
func (sm *SubscriptionModel) Create(sub *Subscription) error {
	ctx := context.Background()

	query := `
		INSERT INTO subscriptions (
			user_id, plan_id, status, 
			current_period_start, current_period_end,
			metadata
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := sm.DB.QueryRow(ctx, query,
		sub.UserID,
		sub.PlanID,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.Metadata,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		slog.Error("Error creating subscription", "error", err, "userID", sub.UserID)
		return err
	}

	slog.Info("Subscription created", "id", sub.ID, "userID", sub.UserID, "plan", sub.PlanID)
	return nil
}

// GetByUserID obtiene la suscripción de un usuario
func (sm *SubscriptionModel) GetByUserID(userID int64) (*Subscription, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, user_id, plan_id, 
			mp_subscription_id, mp_customer_id, mp_preapproval_id,
			status, current_period_start, current_period_end,
			trial_end, canceled_at, cancel_reason,
			metadata, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
	`

	var sub Subscription
	err := sm.DB.QueryRow(ctx, query, userID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.PlanID,
		&sub.MPSubscriptionID,
		&sub.MPCustomerID,
		&sub.MPPreapprovalID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CancelReason,
		&sub.Metadata,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		slog.Error("Error getting subscription by userID", "error", err, "userID", userID)
		return nil, err
	}

	return &sub, nil
}

// GetByID obtiene una suscripción por ID
func (sm *SubscriptionModel) GetByID(id int64) (*Subscription, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, user_id, plan_id, 
			mp_subscription_id, mp_customer_id, mp_preapproval_id,
			status, current_period_start, current_period_end,
			trial_end, canceled_at, cancel_reason,
			metadata, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var sub Subscription
	err := sm.DB.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.PlanID,
		&sub.MPSubscriptionID,
		&sub.MPCustomerID,
		&sub.MPPreapprovalID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CancelReason,
		&sub.Metadata,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		slog.Error("Error getting subscription by ID", "error", err, "id", id)
		return nil, err
	}

	return &sub, nil
}

// GetByMPSubscriptionID obtiene una suscripción por ID de MercadoPago
func (sm *SubscriptionModel) GetByMPSubscriptionID(mpSubID string) (*Subscription, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, user_id, plan_id, 
			mp_subscription_id, mp_customer_id, mp_preapproval_id,
			status, current_period_start, current_period_end,
			trial_end, canceled_at, cancel_reason,
			metadata, created_at, updated_at
		FROM subscriptions
		WHERE mp_subscription_id = $1
	`

	var sub Subscription
	err := sm.DB.QueryRow(ctx, query, mpSubID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.PlanID,
		&sub.MPSubscriptionID,
		&sub.MPCustomerID,
		&sub.MPPreapprovalID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CancelReason,
		&sub.Metadata,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		slog.Error("Error getting subscription by MP ID", "error", err, "mpSubID", mpSubID)
		return nil, err
	}

	return &sub, nil
}

// UpdateStatus actualiza el estado de una suscripción
func (sm *SubscriptionModel) UpdateStatus(id int64, status SubscriptionStatus) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET status = $1
		WHERE id = $2
	`

	result, err := sm.DB.Exec(ctx, query, status, id)
	if err != nil {
		slog.Error("Error updating subscription status", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("Subscription status updated", "id", id, "status", status)
	return nil
}

// UpdateSubscriptionDetails actualiza los detalles de MercadoPago
func (sm *SubscriptionModel) UpdateSubscriptionDetails(id int64, mpSubID, mpCustomerID, mpPreapprovalID string) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET 
			mp_subscription_id = $1,
			mp_customer_id = $2,
			mp_preapproval_id = $3
		WHERE id = $4
	`

	result, err := sm.DB.Exec(ctx, query, mpSubID, mpCustomerID, mpPreapprovalID, id)
	if err != nil {
		slog.Error("Error updating subscription details", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("Subscription details updated", "id", id)
	return nil
}

// UpdatePeriod actualiza el período actual de la suscripción
func (sm *SubscriptionModel) UpdatePeriod(id int64, periodStart, periodEnd time.Time, status SubscriptionStatus) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET 
			current_period_start = $1,
			current_period_end = $2,
			status = $3
		WHERE id = $4
	`

	result, err := sm.DB.Exec(ctx, query, periodStart, periodEnd, status, id)
	if err != nil {
		slog.Error("Error updating subscription period", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("Subscription period updated", "id", id, "periodEnd", periodEnd)
	return nil
}

// Cancel cancela una suscripción
func (sm *SubscriptionModel) Cancel(id int64, reason string) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET 
			status = $1,
			canceled_at = $2,
			cancel_reason = $3
		WHERE id = $4
	`

	now := time.Now()
	result, err := sm.DB.Exec(ctx, query, SubscriptionStatusCanceled, now, reason, id)
	if err != nil {
		slog.Error("Error canceling subscription", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("Subscription canceled", "id", id, "reason", reason)
	return nil
}

// UpdatePlan actualiza el plan de una suscripción
func (sm *SubscriptionModel) UpdatePlan(id int64, newPlanID PlanID) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET plan_id = $1
		WHERE id = $2
	`

	result, err := sm.DB.Exec(ctx, query, newPlanID, id)
	if err != nil {
		slog.Error("Error updating subscription plan", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("Subscription plan updated", "id", id, "newPlan", newPlanID)
	return nil
}

// Activate activa una suscripción
func (sm *SubscriptionModel) Activate(id int64) error {
	return sm.UpdateStatus(id, SubscriptionStatusActive)
}

// UpdateMPPaymentID actualiza el payment_id de MercadoPago
func (sm *SubscriptionModel) UpdateMPPaymentID(id int64, paymentID string) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET mp_subscription_id = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := sm.DB.Exec(ctx, query, paymentID, id)
	if err != nil {
		slog.Error("Error updating MP payment ID", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("MP payment ID updated", "id", id, "paymentID", paymentID)
	return nil
}

// UpdateMPPreapprovalID actualiza el preapproval_id de MercadoPago
func (sm *SubscriptionModel) UpdateMPPreapprovalID(id int64, preapprovalID string) error {
	ctx := context.Background()

	query := `
		UPDATE subscriptions
		SET mp_preapproval_id = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	result, err := sm.DB.Exec(ctx, query, preapprovalID, id)
	if err != nil {
		slog.Error("Error updating MP preapproval ID", "error", err, "id", id)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	slog.Info("MP preapproval ID updated", "id", id, "preapprovalID", preapprovalID)
	return nil
}

// GetExpiringSoon obtiene suscripciones que vencen pronto (para notificaciones)
func (sm *SubscriptionModel) GetExpiringSoon(days int) ([]*Subscription, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, user_id, plan_id, 
			mp_subscription_id, mp_customer_id, mp_preapproval_id,
			status, current_period_start, current_period_end,
			trial_end, canceled_at, cancel_reason,
			metadata, created_at, updated_at
		FROM subscriptions
		WHERE 
			status = 'active' 
			AND current_period_end IS NOT NULL
			AND current_period_end <= NOW() + INTERVAL '1 day' * $1
			AND current_period_end > NOW()
		ORDER BY current_period_end ASC
	`

	rows, err := sm.DB.Query(ctx, query, days)
	if err != nil {
		slog.Error("Error getting expiring subscriptions", "error", err)
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.PlanID,
			&sub.MPSubscriptionID,
			&sub.MPCustomerID,
			&sub.MPPreapprovalID,
			&sub.Status,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&sub.TrialEnd,
			&sub.CanceledAt,
			&sub.CancelReason,
			&sub.Metadata,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			slog.Error("Error scanning expiring subscription", "error", err)
			continue
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, nil
}

// ============================================
// PAYMENT HISTORY MODEL
// ============================================

// PaymentHistoryModel maneja el historial de pagos
type PaymentHistoryModel struct {
	DB *pgxpool.Pool
}

// Create crea un nuevo registro de pago
func (phm *PaymentHistoryModel) Create(payment *PaymentHistory) error {
	ctx := context.Background()

	query := `
		INSERT INTO payment_history (
			subscription_id, user_id, mp_payment_id,
			mp_status, mp_status_detail, amount, currency,
			description, plan_id, payment_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := phm.DB.QueryRow(ctx, query,
		payment.SubscriptionID,
		payment.UserID,
		payment.MPPaymentID,
		payment.MPStatus,
		payment.MPStatusDetail,
		payment.Amount,
		payment.Currency,
		payment.Description,
		payment.PlanID,
		payment.PaymentDate,
	).Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		slog.Error("Error creating payment history", "error", err)
		return err
	}

	slog.Info("Payment history created", "id", payment.ID, "mpPaymentID", payment.MPPaymentID)
	return nil
}

// GetBySubscriptionID obtiene el historial de pagos de una suscripción
func (phm *PaymentHistoryModel) GetBySubscriptionID(subscriptionID int64) ([]*PaymentHistory, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, subscription_id, user_id, mp_payment_id,
			mp_status, mp_status_detail, amount, currency,
			description, plan_id, payment_date, created_at
		FROM payment_history
		WHERE subscription_id = $1
		ORDER BY payment_date DESC
	`

	rows, err := phm.DB.Query(ctx, query, subscriptionID)
	if err != nil {
		slog.Error("Error getting payment history", "error", err)
		return nil, err
	}
	defer rows.Close()

	var payments []*PaymentHistory
	for rows.Next() {
		var p PaymentHistory
		err := rows.Scan(
			&p.ID,
			&p.SubscriptionID,
			&p.UserID,
			&p.MPPaymentID,
			&p.MPStatus,
			&p.MPStatusDetail,
			&p.Amount,
			&p.Currency,
			&p.Description,
			&p.PlanID,
			&p.PaymentDate,
			&p.CreatedAt,
		)
		if err != nil {
			slog.Error("Error scanning payment history", "error", err)
			continue
		}
		payments = append(payments, &p)
	}

	return payments, nil
}

// GetByUserID obtiene el historial de pagos de un usuario
func (phm *PaymentHistoryModel) GetByUserID(userID int64) ([]*PaymentHistory, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id, subscription_id, user_id, mp_payment_id,
			mp_status, mp_status_detail, amount, currency,
			description, plan_id, payment_date, created_at
		FROM payment_history
		WHERE user_id = $1
		ORDER BY payment_date DESC
	`

	rows, err := phm.DB.Query(ctx, query, userID)
	if err != nil {
		slog.Error("Error getting payment history by user", "error", err)
		return nil, err
	}
	defer rows.Close()

	var payments []*PaymentHistory
	for rows.Next() {
		var p PaymentHistory
		err := rows.Scan(
			&p.ID,
			&p.SubscriptionID,
			&p.UserID,
			&p.MPPaymentID,
			&p.MPStatus,
			&p.MPStatusDetail,
			&p.Amount,
			&p.Currency,
			&p.Description,
			&p.PlanID,
			&p.PaymentDate,
			&p.CreatedAt,
		)
		if err != nil {
			slog.Error("Error scanning payment history", "error", err)
			continue
		}
		payments = append(payments, &p)
	}

	return payments, nil
}
