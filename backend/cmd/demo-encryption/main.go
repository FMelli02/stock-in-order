package main

import (
	"fmt"
	"log"
	"time"

	"stock-in-order/backend/internal/crypto"
	"stock-in-order/backend/internal/models"
)

func main() {
	// Ejemplo de uso del sistema de encriptación e integración
	fmt.Println("=== Demo del Sistema de Encriptación de Tokens ===\n")

	// 1. Encriptación básica
	key := "dev-encryption-key-change-me32" // Esta tiene exactamente 32 caracteres
	derivedKey := crypto.DeriveKey(key)     // Pero usamos DeriveKey para asegurar 32 bytes
	tokenMercadoLibre := "APP_USR-1234567890-123456-abcdef1234567890-123456789"

	fmt.Println("1. Encriptación de Token de Mercado Libre:")
	fmt.Printf("   Token original: %s\n", tokenMercadoLibre)

	encrypted, err := crypto.Encrypt(tokenMercadoLibre, derivedKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Token encriptado (bytes): %d bytes\n", len(encrypted))

	decrypted, err := crypto.Decrypt(encrypted, derivedKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Token desencriptado: %s\n", decrypted)
	fmt.Printf("   ✅ Coincide: %v\n\n", tokenMercadoLibre == decrypted)

	// 2. Derivación de clave
	fmt.Println("2. Derivación de Clave:")
	userKey := "mi-clave-super-secreta-123"
	derived := crypto.DeriveKey(userKey)
	fmt.Printf("   Clave original: %s\n", userKey)
	fmt.Printf("   Clave derivada (longitud): %d bytes\n", len(derived))
	fmt.Printf("   ✅ Válida para AES-256\n\n")

	// 3. Ejemplo de estructura de Integration
	fmt.Println("3. Estructura de Integration:")
	integration := models.Integration{
		UserID:         1,
		Platform:       "mercadolibre",
		ExternalUserID: stringPtr("12345678"),
		AccessToken:    tokenMercadoLibre,
		RefreshToken:   "TG-abc123def456",
		ExpiresAt:      time.Now().Add(6 * time.Hour),
		CreatedAt:      time.Now(),
	}

	fmt.Printf("   User ID: %d\n", integration.UserID)
	fmt.Printf("   Platform: %s\n", integration.Platform)
	fmt.Printf("   External User ID: %s\n", *integration.ExternalUserID)
	fmt.Printf("   Access Token: %s... (truncado)\n", integration.AccessToken[:20])
	fmt.Printf("   Expires At: %s\n", integration.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("   Token Expired: %v\n\n", integration.IsTokenExpired())

	// 4. Seguridad
	fmt.Println("4. Notas de Seguridad:")
	fmt.Println("   ✅ Los tokens se encriptan con AES-256-GCM")
	fmt.Println("   ✅ Cada encriptación usa un nonce aleatorio único")
	fmt.Println("   ✅ Los tokens nunca se almacenan en texto plano")
	fmt.Println("   ✅ La clave de encriptación debe tener 32 bytes (256 bits)")
	fmt.Println("   ✅ En producción, usar una clave aleatoria fuerte")
	fmt.Println("\n=== Fin de la Demo ===")
}

func stringPtr(s string) *string {
	return &s
}
