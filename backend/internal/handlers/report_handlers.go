package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"

	"stock-in-order/backend/internal/middleware"
	"stock-in-order/backend/internal/models"
	"stock-in-order/backend/internal/rabbitmq"
)

// ExportProductsXLSX maneja GET /api/v1/reports/products/xlsx
// Genera un archivo Excel profesional con todos los productos del usuario
func ExportProductsXLSX(db *pgxpool.Pool) http.HandlerFunc {
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

		// Crear un nuevo archivo Excel en memoria
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				// Log error if needed
			}
		}()

		// Crear una nueva hoja llamada "Productos"
		sheetName := "Productos"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			http.Error(w, "could not create Excel sheet", http.StatusInternalServerError)
			return
		}

		// Establecer la hoja activa
		f.SetActiveSheet(index)

		// Escribir cabeceras en la fila 1
		headers := []string{"ID", "Nombre", "SKU", "Descripción", "Cantidad", "Fecha de Creación"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Escribir filas de datos (a partir de la fila 2)
		for rowIndex, product := range products {
			description := ""
			if product.Description != nil {
				description = *product.Description
			}

			row := rowIndex + 2 // Comenzar desde la fila 2 (después de headers)

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), product.ID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), product.Name)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), product.SKU)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), description)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), product.Quantity)
			f.SetCellValue(sheetName, "F"+strconv.Itoa(row), product.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		// Configurar headers HTTP para descarga de archivo Excel
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\"productos.xlsx\"")

		// Escribir el archivo Excel en el ResponseWriter
		if err := f.Write(w); err != nil {
			http.Error(w, "could not write Excel file", http.StatusInternalServerError)
			return
		}
	}
}

// ExportCustomersXLSX maneja GET /api/v1/reports/customers/xlsx
// Genera un archivo Excel profesional con todos los clientes del usuario
func ExportCustomersXLSX(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener todos los clientes del usuario
		cm := &models.CustomerModel{DB: db}
		customers, err := cm.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch customers", http.StatusInternalServerError)
			return
		}

		// Crear un nuevo archivo Excel en memoria
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				// Log error if needed
			}
		}()

		// Crear una nueva hoja llamada "Clientes"
		sheetName := "Clientes"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			http.Error(w, "could not create Excel sheet", http.StatusInternalServerError)
			return
		}

		// Establecer la hoja activa
		f.SetActiveSheet(index)

		// Escribir cabeceras en la fila 1
		headers := []string{"ID", "Nombre", "Email", "Teléfono", "Dirección", "Fecha de Creación"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Escribir filas de datos (a partir de la fila 2)
		for rowIndex, customer := range customers {
			row := rowIndex + 2

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), customer.ID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), customer.Name)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), customer.Email)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), customer.Phone)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), customer.Address)
			f.SetCellValue(sheetName, "F"+strconv.Itoa(row), customer.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		// Configurar headers HTTP para descarga de archivo Excel
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\"clientes.xlsx\"")

		// Escribir el archivo Excel en el ResponseWriter
		if err := f.Write(w); err != nil {
			http.Error(w, "could not write Excel file", http.StatusInternalServerError)
			return
		}
	}
}

// ExportSuppliersXLSX maneja GET /api/v1/reports/suppliers/xlsx
// Genera un archivo Excel profesional con todos los proveedores del usuario
func ExportSuppliersXLSX(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener todos los proveedores del usuario
		sm := &models.SupplierModel{DB: db}
		suppliers, err := sm.GetAllForUser(userID)
		if err != nil {
			http.Error(w, "could not fetch suppliers", http.StatusInternalServerError)
			return
		}

		// Crear un nuevo archivo Excel en memoria
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				// Log error if needed
			}
		}()

		// Crear una nueva hoja llamada "Proveedores"
		sheetName := "Proveedores"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			http.Error(w, "could not create Excel sheet", http.StatusInternalServerError)
			return
		}

		// Establecer la hoja activa
		f.SetActiveSheet(index)

		// Escribir cabeceras en la fila 1
		headers := []string{"ID", "Nombre", "Email", "Teléfono", "Dirección", "Fecha de Creación"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Escribir filas de datos (a partir de la fila 2)
		for rowIndex, supplier := range suppliers {
			row := rowIndex + 2

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), supplier.ID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), supplier.Name)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), supplier.Email)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), supplier.Phone)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), supplier.Address)
			f.SetCellValue(sheetName, "F"+strconv.Itoa(row), supplier.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		// Configurar headers HTTP para descarga de archivo Excel
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\"proveedores.xlsx\"")

		// Escribir el archivo Excel en el ResponseWriter
		if err := f.Write(w); err != nil {
			http.Error(w, "could not write Excel file", http.StatusInternalServerError)
			return
		}
	}
}

// ExportSalesOrdersXLSX maneja GET /api/v1/reports/sales-orders/xlsx
// Genera un archivo Excel con órdenes de venta filtradas por fecha y estado
func ExportSalesOrdersXLSX(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Leer filtros de query params
		filters := models.SalesOrderFilters{
			DateFrom: r.URL.Query().Get("date_from"),
			DateTo:   r.URL.Query().Get("date_to"),
			Status:   r.URL.Query().Get("status"),
		}

		// Obtener órdenes filtradas
		som := &models.SalesOrderModel{DB: db}
		orders, err := som.GetAllForUserWithFilters(userID, filters)
		if err != nil {
			http.Error(w, "could not fetch sales orders", http.StatusInternalServerError)
			return
		}

		// Crear archivo Excel
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				// Log error if needed
			}
		}()

		sheetName := "Ventas"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			http.Error(w, "could not create Excel sheet", http.StatusInternalServerError)
			return
		}

		f.SetActiveSheet(index)

		// Headers
		headers := []string{"ID", "Cliente", "Fecha", "Estado", "Total"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Datos
		for rowIndex, order := range orders {
			row := rowIndex + 2

			totalAmount := 0.0
			if order.TotalAmount.Valid {
				totalAmount = order.TotalAmount.Float64
			}

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), order.ID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), order.CustomerName)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), order.OrderDate.Format("2006-01-02"))
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), order.Status)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), totalAmount)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\"ventas.xlsx\"")

		if err := f.Write(w); err != nil {
			http.Error(w, "could not write Excel file", http.StatusInternalServerError)
			return
		}
	}
}

// ExportPurchaseOrdersXLSX maneja GET /api/v1/reports/purchase-orders/xlsx
// Genera un archivo Excel con órdenes de compra filtradas por fecha y estado
func ExportPurchaseOrdersXLSX(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Leer filtros de query params
		filters := models.PurchaseOrderFilters{
			DateFrom: r.URL.Query().Get("date_from"),
			DateTo:   r.URL.Query().Get("date_to"),
			Status:   r.URL.Query().Get("status"),
		}

		// Obtener órdenes filtradas
		pom := &models.PurchaseOrderModel{DB: db}
		orders, err := pom.GetAllForUserWithFilters(userID, filters)
		if err != nil {
			http.Error(w, "could not fetch purchase orders", http.StatusInternalServerError)
			return
		}

		// Crear archivo Excel
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				// Log error if needed
			}
		}()

		sheetName := "Compras"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			http.Error(w, "could not create Excel sheet", http.StatusInternalServerError)
			return
		}

		f.SetActiveSheet(index)

		// Headers
		headers := []string{"ID", "Proveedor", "Fecha", "Estado"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Datos
		for rowIndex, order := range orders {
			row := rowIndex + 2

			orderDate := ""
			if order.OrderDate != nil {
				orderDate = order.OrderDate.Format("2006-01-02")
			}

			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), order.ID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), order.SupplierName)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), orderDate)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), order.Status)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\"compras.xlsx\"")

		if err := f.Write(w); err != nil {
			http.Error(w, "could not write Excel file", http.StatusInternalServerError)
			return
		}
	}
}

// ============================================================================
// NUEVOS HANDLERS ASÍNCRONOS - DELEGACIÓN AL WORKER
// ============================================================================

// RequestProductsReportByEmail maneja POST /api/v1/reports/products/email
// Publica un mensaje a RabbitMQ para que el worker genere el reporte y lo envíe por email
func RequestProductsReportByEmail(db *pgxpool.Pool, rabbit *rabbitmq.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener el email del usuario desde la base de datos
		um := &models.UserModel{DB: db}
		user, err := um.GetByID(userID)
		if err != nil {
			http.Error(w, "could not fetch user info", http.StatusInternalServerError)
			return
		}

		// Crear mensaje para RabbitMQ
		req := rabbitmq.ReportRequest{
			UserID:     userID,
			Email:      user.Email,
			ReportType: "products",
		}

		// Publicar mensaje a la cola
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := rabbit.PublishReportRequest(ctx, req); err != nil {
			http.Error(w, "could not queue report request", http.StatusInternalServerError)
			return
		}

		// Responder inmediatamente con 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Tu reporte se está generando y te llegará por email en unos minutos.",
		})
	}
}

// RequestCustomersReportByEmail maneja POST /api/v1/reports/customers/email
// Publica un mensaje a RabbitMQ para que el worker genere el reporte y lo envíe por email
func RequestCustomersReportByEmail(db *pgxpool.Pool, rabbit *rabbitmq.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener el email del usuario desde la base de datos
		um := &models.UserModel{DB: db}
		user, err := um.GetByID(userID)
		if err != nil {
			http.Error(w, "could not fetch user info", http.StatusInternalServerError)
			return
		}

		// Crear mensaje para RabbitMQ
		req := rabbitmq.ReportRequest{
			UserID:     userID,
			Email:      user.Email,
			ReportType: "customers",
		}

		// Publicar mensaje a la cola
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := rabbit.PublishReportRequest(ctx, req); err != nil {
			http.Error(w, "could not queue report request", http.StatusInternalServerError)
			return
		}

		// Responder inmediatamente con 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Tu reporte se está generando y te llegará por email en unos minutos.",
		})
	}
}

// RequestSuppliersReportByEmail maneja POST /api/v1/reports/suppliers/email
// Publica un mensaje a RabbitMQ para que el worker genere el reporte y lo envíe por email
func RequestSuppliersReportByEmail(db *pgxpool.Pool, rabbit *rabbitmq.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Obtener el email del usuario desde la base de datos
		um := &models.UserModel{DB: db}
		user, err := um.GetByID(userID)
		if err != nil {
			http.Error(w, "could not fetch user info", http.StatusInternalServerError)
			return
		}

		// Crear mensaje para RabbitMQ
		req := rabbitmq.ReportRequest{
			UserID:     userID,
			Email:      user.Email,
			ReportType: "suppliers",
		}

		// Publicar mensaje a la cola
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := rabbit.PublishReportRequest(ctx, req); err != nil {
			http.Error(w, "could not queue report request", http.StatusInternalServerError)
			return
		}

		// Responder inmediatamente con 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Tu reporte se está generando y te llegará por email en unos minutos.",
		})
	}
}
