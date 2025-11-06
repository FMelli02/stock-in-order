package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/models"
)

// ============================================
// PLAN LIMITS ENFORCEMENT - "EL PATOVICA 3.0"
// ============================================

// RequireFeature verifica que el plan del usuario tenga acceso a una feature específica
func RequireFeature(db *pgxpool.Pool, feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener suscripción del contexto (inyectada por RequireActiveSubscription)
			subscription, ok := SubscriptionFromContext(r.Context())
			if !ok {
				// Fallback: obtener desde DB
				userID, ok := UserIDFromContext(r.Context())
				if !ok {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				sm := &models.SubscriptionModel{DB: db}
				var err error
				subscription, err = sm.GetByUserID(userID)
				if err != nil {
					slog.Error("Error obteniendo suscripción", "error", err)
					http.Error(w, "Error verificando permisos", http.StatusInternalServerError)
					return
				}
			}

			// Obtener features del plan
			features := models.GetPlanFeatures(subscription.PlanID)

			// Verificar la feature específica
			hasFeature := false
			switch feature {
			case "reports":
				hasFeature = features.AdvancedReports
			case "email_notifications":
				hasFeature = features.EmailNotifications
			case "api_access":
				hasFeature = features.APIAccess
			case "multi_warehouse":
				hasFeature = features.MultiWarehouse
			case "batch_tracking":
				hasFeature = features.BatchTracking
			case "expiry_management":
				hasFeature = features.ExpiryManagement
			case "priority_support":
				hasFeature = features.PrioritySupport
			case "custom_integrations":
				hasFeature = features.CustomIntegrations
			default:
				// Feature desconocida → denegar por defecto
				hasFeature = false
			}

			if !hasFeature {
				slog.Info("Acceso denegado: feature no disponible en el plan",
					"userID", subscription.UserID,
					"plan", subscription.PlanID,
					"feature", feature)

				w.WriteHeader(http.StatusPaymentRequired)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":       "Feature no disponible en tu plan",
					"message":     "Tu plan " + string(subscription.PlanID) + " no incluye acceso a " + feature,
					"feature":     feature,
					"plan":        subscription.PlanID,
					"action":      "upgrade",
					"upgrade_url": "/subscriptions/status",
				})
				return
			}

			// Feature disponible → permitir acceso
			next.ServeHTTP(w, r)
		})
	}
}

// CheckProductLimit verifica si el usuario puede crear más productos
func CheckProductLimit(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Solo aplicar a POST (creación)
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Obtener suscripción
			sm := &models.SubscriptionModel{DB: db}
			subscription, err := sm.GetByUserID(userID)
			if err != nil {
				slog.Error("Error obteniendo suscripción", "error", err)
				http.Error(w, "Error verificando límites", http.StatusInternalServerError)
				return
			}

			// Obtener límites del plan
			features := models.GetPlanFeatures(subscription.PlanID)

			// Si es ilimitado (-1), permitir
			if features.MaxProducts == -1 {
				next.ServeHTTP(w, r)
				return
			}

			// TODO: Implementar conteo real de productos por usuario
			// Por ahora, no bloqueamos la creación (placeholder)
			// En producción, agregar método CountByUser al ProductModel

			// Placeholder: siempre permitir
			// currentCount, err := pm.CountByUser(userID)
			// if currentCount >= features.MaxProducts { ... }

			next.ServeHTTP(w, r)
		})
	}
}

// CheckOrderLimit verifica si el usuario puede crear más órdenes este mes
func CheckOrderLimit(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Solo aplicar a POST (creación)
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Obtener suscripción
			sm := &models.SubscriptionModel{DB: db}
			subscription, err := sm.GetByUserID(userID)
			if err != nil {
				slog.Error("Error obteniendo suscripción", "error", err)
				http.Error(w, "Error verificando límites", http.StatusInternalServerError)
				return
			}

			// Obtener límites del plan
			features := models.GetPlanFeatures(subscription.PlanID)

			// Si es ilimitado (-1), permitir
			if features.MaxOrders == -1 {
				next.ServeHTTP(w, r)
				return
			}

			// Contar órdenes del mes actual
			// Aquí necesitaríamos un método CountOrdersThisMonth en el modelo
			// Por ahora, simplificamos asumiendo que todas las órdenes cuentan
			// TODO: Implementar conteo por mes

			// Por ahora, permitir siempre (placeholder)
			// En producción, implementar conteo real
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================
// HELPERS
// ============================================

// formatNumber formatea un número para mostrar en mensajes
func formatNumber(n int) string {
	if n == -1 {
		return "ilimitados"
	}
	// Aquí podríamos agregar formato con separadores de miles
	return string(rune(n + '0'))
}
