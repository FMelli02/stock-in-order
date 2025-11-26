package reports

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/xuri/excelize/v2"

	"stock-in-order/worker/internal/models"
)

// GenerateProductsReport genera un reporte Excel de productos para un usuario
// Retorna el archivo Excel como un slice de bytes
func GenerateProductsReport(db *pgxpool.Pool, userID int64) ([]byte, error) {
	// Obtener todos los productos del usuario
	pm := &models.ProductModel{DB: db}
	products, err := pm.GetAllForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch products: %w", err)
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
		return nil, fmt.Errorf("could not create Excel sheet: %w", err)
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

		row := rowIndex + 2 // Comenzar desde la fila 2

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), product.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), product.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), product.SKU)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), description)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), product.Quantity)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), product.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Escribir el archivo Excel a un buffer en memoria
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("could not write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateCustomersReport genera un reporte Excel de clientes para un usuario
func GenerateCustomersReport(db *pgxpool.Pool, userID int64) ([]byte, error) {
	cm := &models.CustomerModel{DB: db}
	customers, err := cm.GetAllForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch customers: %w", err)
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error if needed
		}
	}()

	sheetName := "Clientes"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("could not create Excel sheet: %w", err)
	}

	f.SetActiveSheet(index)

	headers := []string{"ID", "Nombre", "Email", "Teléfono", "Dirección", "Fecha de Creación"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for rowIndex, customer := range customers {
		row := rowIndex + 2

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), customer.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), customer.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), customer.Email)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), customer.Phone)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), customer.Address)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), customer.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("could not write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateSuppliersReport genera un reporte Excel de proveedores para un usuario
func GenerateSuppliersReport(db *pgxpool.Pool, userID int64) ([]byte, error) {
	sm := &models.SupplierModel{DB: db}
	suppliers, err := sm.GetAllForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch suppliers: %w", err)
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error if needed
		}
	}()

	sheetName := "Proveedores"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("could not create Excel sheet: %w", err)
	}

	f.SetActiveSheet(index)

	headers := []string{"ID", "Nombre", "Email", "Teléfono", "Dirección", "Fecha de Creación"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for rowIndex, supplier := range suppliers {
		row := rowIndex + 2

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), supplier.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), supplier.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), supplier.Email)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), supplier.Phone)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), supplier.Address)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), supplier.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("could not write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateSalesOrderPDF genera un PDF de comprobante para una orden de venta
func GenerateSalesOrderPDF(db *pgxpool.Pool, orderID int64, userID int64) ([]byte, error) {
	ctx := context.Background()

	// Obtener datos de la orden de venta con items
	const qOrder = `
		SELECT 
			so.id, so.order_date, so.status, so.total_amount,
			COALESCE(c.name, 'Cliente No Especificado') AS customer_name
		FROM sales_orders so
		LEFT JOIN customers c ON so.customer_id = c.id
		WHERE so.id = $1 AND so.user_id = $2`

	var order models.SalesOrderWithItems
	err := db.QueryRow(ctx, qOrder, orderID, userID).
		Scan(&order.ID, &order.OrderDate, &order.Status, &order.TotalAmount, &order.CustomerName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch order: %w", err)
	}

	// Obtener items con nombre de producto
	const qItems = `
		SELECT 
			p.name, oi.quantity, oi.unit_price, (oi.quantity * oi.unit_price) AS subtotal
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.id`

	rows, err := db.Query(ctx, qItems, orderID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch items: %w", err)
	}
	defer rows.Close()

	order.Items = []models.OrderItemPDF{}
	for rows.Next() {
		var item models.OrderItemPDF
		if err := rows.Scan(&item.ProductName, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, fmt.Errorf("could not scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating items: %w", rows.Err())
	}

	// Crear PDF con maroto
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(20, 10, 20)

	// Header con color de fondo
	m.SetBackgroundColor(color.Color{Red: 55, Green: 65, Blue: 81}) // gray-700
	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("COMPROBANTE DE VENTA", props.Text{
				Top:   2,
				Size:  16,
				Align: consts.Center,
				Style: consts.Bold,
				Color: color.Color{Red: 255, Green: 255, Blue: 255},
			})
		})
	})
	m.SetBackgroundColor(color.NewWhite())

	// Información de la orden
	m.Row(8, func() {
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Orden N°: %d", order.ID), props.Text{
				Top:   2,
				Size:  12,
				Style: consts.Bold,
			})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Fecha: %s", order.OrderDate.Format("02/01/2006")), props.Text{
				Top:   2,
				Size:  10,
				Align: consts.Right,
			})
		})
	})

	// Información del cliente
	m.Row(8, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Cliente: %s", order.CustomerName), props.Text{
				Top:  2,
				Size: 11,
			})
		})
	})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Estado: %s", order.Status), props.Text{
				Top:  1,
				Size: 10,
			})
		})
	})

	m.Line(2.0)

	// Tabla de items - Encabezado
	m.SetBackgroundColor(color.Color{Red: 229, Green: 231, Blue: 235}) // gray-200
	m.Row(8, func() {
		m.Col(5, func() {
			m.Text("Producto", props.Text{
				Top:   2,
				Size:  10,
				Style: consts.Bold,
				Align: consts.Left,
			})
		})
		m.Col(2, func() {
			m.Text("Cantidad", props.Text{
				Top:   2,
				Size:  10,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
		m.Col(2, func() {
			m.Text("Precio Unit.", props.Text{
				Top:   2,
				Size:  10,
				Style: consts.Bold,
				Align: consts.Right,
			})
		})
		m.Col(3, func() {
			m.Text("Subtotal", props.Text{
				Top:   2,
				Size:  10,
				Style: consts.Bold,
				Align: consts.Right,
			})
		})
	})
	m.SetBackgroundColor(color.NewWhite())

	// Tabla de items - Filas
	for _, item := range order.Items {
		m.Row(7, func() {
			m.Col(5, func() {
				m.Text(item.ProductName, props.Text{
					Top:  1.5,
					Size: 9,
				})
			})
			m.Col(2, func() {
				m.Text(strconv.Itoa(item.Quantity), props.Text{
					Top:   1.5,
					Size:  9,
					Align: consts.Center,
				})
			})
			m.Col(2, func() {
				m.Text(fmt.Sprintf("$%.2f", item.UnitPrice), props.Text{
					Top:   1.5,
					Size:  9,
					Align: consts.Right,
				})
			})
			m.Col(3, func() {
				m.Text(fmt.Sprintf("$%.2f", item.Subtotal), props.Text{
					Top:   1.5,
					Size:  9,
					Align: consts.Right,
				})
			})
		})
	}

	m.Line(1.0)

	// Total
	m.Row(10, func() {
		m.Col(9, func() {
			m.Text("TOTAL:", props.Text{
				Top:   3,
				Size:  12,
				Style: consts.Bold,
				Align: consts.Right,
			})
		})
		m.Col(3, func() {
			m.Text(fmt.Sprintf("$%.2f", order.TotalAmount), props.Text{
				Top:   3,
				Size:  12,
				Style: consts.Bold,
				Align: consts.Right,
			})
		})
	})

	// Footer
	m.Row(15, func() {
		m.Col(12, func() {
			m.Text("Gracias por su compra", props.Text{
				Top:   8,
				Size:  10,
				Align: consts.Center,
				Style: consts.Italic,
			})
		})
	})

	// Generar el PDF y obtener los bytes
	pdfBytes, err := m.Output()
	if err != nil {
		return nil, fmt.Errorf("could not generate PDF: %w", err)
	}

	return pdfBytes.Bytes(), nil
}
