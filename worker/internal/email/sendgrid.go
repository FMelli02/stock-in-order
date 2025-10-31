package email

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Client encapsula la configuración de SendGrid
type Client struct {
	apiKey     string
	fromEmail  string
	fromName   string
	sgClient   *sendgrid.Client
	isDisabled bool // Para desarrollo sin SendGrid configurado
}

// NewClient crea un nuevo cliente de SendGrid
func NewClient(apiKey, fromEmail, fromName string) *Client {
	if apiKey == "" {
		log.Println("⚠️  SENDGRID_API_KEY no configurado. Los emails NO se enviarán.")
		return &Client{
			isDisabled: true,
		}
	}

	return &Client{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		fromName:   fromName,
		sgClient:   sendgrid.NewSendClient(apiKey),
		isDisabled: false,
	}
}

// EmailAttachment representa un archivo adjunto
type EmailAttachment struct {
	Filename    string // Nombre del archivo (ej: "reporte_productos.xlsx")
	Content     []byte // Contenido del archivo en bytes
	ContentType string // MIME type (ej: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
}

// SendReportEmail envía un email con el reporte en Excel adjunto
func (c *Client) SendReportEmail(toEmail, toName, reportType string, attachment EmailAttachment) error {
	if c.isDisabled {
		log.Printf("📧 [MODO DEV] Email simulado a %s - Adjunto: %s (%d bytes)", toEmail, attachment.Filename, len(attachment.Content))
		return nil
	}

	// Crear el email desde
	from := mail.NewEmail(c.fromName, c.fromEmail)

	// Crear el email hacia
	to := mail.NewEmail(toName, toEmail)

	// Asunto y contenido según el tipo de reporte
	subject, htmlContent := getEmailContent(reportType)

	// Crear el mensaje
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	// Adjuntar el archivo Excel (convertir a base64)
	a := mail.NewAttachment()
	encoded := base64.StdEncoding.EncodeToString(attachment.Content)
	a.SetContent(encoded)
	a.SetType(attachment.ContentType)
	a.SetFilename(attachment.Filename)
	a.SetDisposition("attachment")
	message.AddAttachment(a)

	// Enviar el email
	response, err := c.sgClient.Send(message)
	if err != nil {
		return fmt.Errorf("error al enviar email: %w", err)
	}

	// Verificar respuesta
	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid respondió con código %d: %s", response.StatusCode, response.Body)
	}

	log.Printf("✅ Email enviado exitosamente a %s (código: %d)", toEmail, response.StatusCode)
	return nil
}

// SendStockAlertEmail envía un email de alerta de stock bajo
func (c *Client) SendStockAlertEmail(toEmail, productName string, currentStock, minStock int) error {
	if c.isDisabled {
		log.Printf("📧 [MODO DEV] Alerta de stock simulada a %s - Producto: %s (%d/%d)", toEmail, productName, currentStock, minStock)
		return nil
	}

	// Crear el email desde
	from := mail.NewEmail(c.fromName, c.fromEmail)

	// Crear el email hacia
	to := mail.NewEmail("", toEmail)

	// Asunto
	subject := "⚠️ ALERTA DE STOCK BAJO - " + productName

	// Contenido HTML
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #fff9f9; padding: 30px; border-radius: 0 0 10px 10px; border: 2px solid #ff6b6b; }
        .alert-box { background: #ffe0e0; border-left: 4px solid #ff6b6b; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .stock-info { background: white; padding: 15px; border-radius: 8px; margin: 15px 0; }
        .stock-current { font-size: 36px; font-weight: bold; color: #ff6b6b; }
        .stock-min { font-size: 24px; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .warning-icon { font-size: 48px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="warning-icon">⚠️</div>
            <h1>ALERTA DE STOCK BAJO</h1>
        </div>
        <div class="content">
            <p><strong>¡OJO!</strong> Te estás quedando sin el siguiente producto:</p>
            
            <div class="alert-box">
                <h2 style="margin-top: 0; color: #ff6b6b;">%s</h2>
            </div>
            
            <div class="stock-info">
                <p style="margin: 5px 0; color: #666;">Stock Actual:</p>
                <div class="stock-current">%d unidades</div>
                
                <p style="margin: 15px 0 5px 0; color: #666;">Tu Stock Mínimo:</p>
                <div class="stock-min">%d unidades</div>
            </div>
            
            <div class="alert-box">
                <p style="margin: 0;"><strong>⚡ Acción Requerida:</strong></p>
                <p style="margin: 5px 0 0 0;">Te recomendamos realizar un nuevo pedido a tu proveedor lo antes posible para evitar quedarte sin stock.</p>
            </div>
            
            <p style="margin-top: 20px;">Puedes gestionar tus productos desde el panel de <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático de alerta de stock.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`, productName, currentStock, minStock)

	// Crear el mensaje
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	// Enviar el email
	response, err := c.sgClient.Send(message)
	if err != nil {
		return fmt.Errorf("error al enviar email de alerta: %w", err)
	}

	// Verificar respuesta
	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid respondió con código %d: %s", response.StatusCode, response.Body)
	}

	log.Printf("✅ Alerta de stock enviada a %s (código: %d)", toEmail, response.StatusCode)
	return nil
}

// getEmailContent devuelve el asunto y contenido HTML según el tipo de reporte
func getEmailContent(reportType string) (subject string, htmlContent string) {
	switch reportType {
	case "products":
		subject = "📦 Tu Reporte de Productos está Listo"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin-top: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📦 Reporte de Productos</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu reporte de <strong>Productos</strong> ha sido generado exitosamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información actualizada de tu inventario.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Código y nombre de productos</li>
                <li>✅ Descripción y categoría</li>
                <li>✅ Precios actualizados</li>
                <li>✅ Stock disponible</li>
                <li>✅ Fechas de registro</li>
            </ul>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	case "products_weekly":
		subject = "📦 Reporte Semanal de Productos"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin-top: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📦 Reporte Semanal de Productos</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu <strong>reporte semanal de Productos</strong> ha sido generado automáticamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información actualizada de tu inventario.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Código y nombre de productos</li>
                <li>✅ Descripción y categoría</li>
                <li>✅ Precios actualizados</li>
                <li>✅ Stock disponible</li>
                <li>✅ Fechas de registro</li>
            </ul>
            <p>Este reporte se genera automáticamente cada semana.</p>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	case "customers":
		subject = "👥 Tu Reporte de Clientes está Listo"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 Reporte de Clientes</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu reporte de <strong>Clientes</strong> ha sido generado exitosamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información de tu cartera de clientes.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Nombre completo del cliente</li>
                <li>✅ Email y teléfono</li>
                <li>✅ Dirección completa</li>
                <li>✅ Fecha de registro</li>
            </ul>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	case "customers_weekly":
		subject = "👥 Reporte Semanal de Clientes"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 Reporte Semanal de Clientes</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu <strong>reporte semanal de Clientes</strong> ha sido generado automáticamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información de tu cartera de clientes.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Nombre completo del cliente</li>
                <li>✅ Email y teléfono</li>
                <li>✅ Dirección completa</li>
                <li>✅ Fecha de registro</li>
            </ul>
            <p>Este reporte se genera automáticamente cada semana.</p>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	case "suppliers":
		subject = "🏭 Tu Reporte de Proveedores está Listo"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏭 Reporte de Proveedores</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu reporte de <strong>Proveedores</strong> ha sido generado exitosamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información de tus proveedores.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Nombre del proveedor</li>
                <li>✅ Contacto y email</li>
                <li>✅ Teléfono</li>
                <li>✅ Dirección completa</li>
                <li>✅ Fecha de registro</li>
            </ul>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	case "suppliers_weekly":
		subject = "🏭 Reporte Semanal de Proveedores"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏭 Reporte Semanal de Proveedores</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu <strong>reporte semanal de Proveedores</strong> ha sido generado automáticamente y está adjunto en este email.</p>
            <p>El archivo está en formato Excel (.xlsx) y contiene toda la información de tus proveedores.</p>
            <p><strong>¿Qué incluye el reporte?</strong></p>
            <ul>
                <li>✅ Nombre del proveedor</li>
                <li>✅ Contacto y email</li>
                <li>✅ Teléfono</li>
                <li>✅ Dirección completa</li>
                <li>✅ Fecha de registro</li>
            </ul>
            <p>Este reporte se genera automáticamente cada semana.</p>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	default:
		subject = "📊 Tu Reporte está Listo"
		htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Reporte Generado</h1>
        </div>
        <div class="content">
            <p>¡Hola!</p>
            <p>Tu reporte ha sido generado exitosamente y está adjunto en este email.</p>
            <p>Gracias por usar <strong>Stock in Order</strong>.</p>
        </div>
        <div class="footer">
            <p>Este es un email automático, por favor no responder.</p>
            <p>Stock in Order &copy; 2025</p>
        </div>
    </div>
</body>
</html>
`
	}
	return
}
