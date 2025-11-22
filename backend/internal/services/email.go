package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService handles email sending via SendGrid
type EmailService struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		apiKey:    os.Getenv("SENDGRID_API_KEY"),
		fromEmail: getEnvOrDefault("SENDGRID_FROM_EMAIL", "noreply@stock-in-order.com"),
		fromName:  getEnvOrDefault("SENDGRID_FROM_NAME", "Stock In Order"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SendPasswordResetEmail sends a password reset email with a reset link
func (s *EmailService) SendPasswordResetEmail(toEmail string, data map[string]interface{}) error {
	if s.apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY not configured")
	}

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("", toEmail)

	subject := "Recuperación de contraseña - Stock In Order"

	// Extract data
	userName, _ := data["user_name"].(string)
	resetLink, _ := data["reset_link"].(string)
	expiryHours, _ := data["expiry_hours"].(int)

	// HTML email content
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px;">
        <h1 style="color: #4F46E5; margin-top: 0;">Recuperación de Contraseña</h1>
        
        <p>Hola%s,</p>
        
        <p>Recibimos una solicitud para restablecer tu contraseña en <strong>Stock In Order</strong>.</p>
        
        <p>Para crear una nueva contraseña, haz clic en el siguiente botón:</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" 
               style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
                Restablecer Contraseña
            </a>
        </div>
        
        <p>O copia y pega este enlace en tu navegador:</p>
        <p style="background-color: #e9ecef; padding: 10px; border-radius: 3px; word-break: break-all; font-size: 12px;">
            %s
        </p>
        
        <p style="color: #dc3545; font-weight: bold;">⚠️ Este enlace expirará en %d hora(s).</p>
        
        <hr style="border: none; border-top: 1px solid #dee2e6; margin: 20px 0;">
        
        <p style="font-size: 14px; color: #6c757d;">
            <strong>¿No solicitaste este cambio?</strong><br>
            Si no solicitaste restablecer tu contraseña, puedes ignorar este correo de forma segura. Tu contraseña actual no será modificada.
        </p>
        
        <p style="font-size: 12px; color: #6c757d; margin-top: 20px;">
            Saludos,<br>
            El equipo de <strong>Stock In Order</strong>
        </p>
    </div>
</body>
</html>
`, getUserNameGreeting(userName), resetLink, resetLink, expiryHours)

	// Plain text fallback
	plainTextContent := fmt.Sprintf(`
Recuperación de Contraseña - Stock In Order

Hola%s,

Recibimos una solicitud para restablecer tu contraseña en Stock In Order.

Para crear una nueva contraseña, visita el siguiente enlace:
%s

⚠️ Este enlace expirará en %d hora(s).

¿No solicitaste este cambio?
Si no solicitaste restablecer tu contraseña, puedes ignorar este correo de forma segura.

Saludos,
El equipo de Stock In Order
`, getUserNameGreeting(userName), resetLink, expiryHours)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid returned error status: %d - %s", response.StatusCode, response.Body)
	}

	return nil
}

func getUserNameGreeting(userName string) string {
	if userName != "" {
		return " " + userName
	}
	return ""
}
