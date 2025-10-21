package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
)

// ExportProductsCSV maneja GET /api/v1/reports/products/csv
// Genera un archivo CSV con todos los productos del usuario
func ExportProductsCSV(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener todos los productos del usuario
		pm := &models.ProductModel{DB: db}
		products, err := pm.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch products", http.StatusInternalServerError)
			return
		}

		// Configurar headers para descarga de CSV
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"productos.csv\"")

		// Crear writer CSV
		writer := csv.NewWriter(w)
		defer writer.Flush()

		// Escribir cabeceras
		headers := []string{"ID", "Nombre", "SKU", "Descripción", "Cantidad", "Fecha de Creación"}
		if err := writer.Write(headers); err != nil {
			http.Error(w, "could not write CSV headers", http.StatusInternalServerError)
			return
		}

		// Escribir filas de datos
		for _, product := range products {
			description := ""
			if product.Description != nil {
				description = *product.Description
			}

			row := []string{
				strconv.FormatInt(product.ID, 10),
				product.Name,
				product.SKU,
				description,
				strconv.Itoa(product.Quantity),
				product.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			if err := writer.Write(row); err != nil {
				// Si ya comenzamos a escribir, no podemos enviar error HTTP
				return
			}
		}
	}
}
