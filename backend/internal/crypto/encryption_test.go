package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	plaintext := "Este es un token secreto de Mercado Libre"

	// Encriptar
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Error al encriptar: %v", err)
	}

	// Verificar que el texto cifrado no sea igual al texto plano
	if string(ciphertext) == plaintext {
		t.Fatal("El texto cifrado es igual al texto plano")
	}

	// Desencriptar
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Error al desencriptar: %v", err)
	}

	// Verificar que el texto desencriptado sea igual al original
	if decrypted != plaintext {
		t.Fatalf("Texto desencriptado no coincide. Esperado: %s, Obtenido: %s", plaintext, decrypted)
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	invalidKey := "clave-muy-corta" // Menos de 32 bytes
	plaintext := "test"

	_, err := Encrypt(plaintext, invalidKey)
	if err != ErrInvalidKey {
		t.Fatalf("Esperaba ErrInvalidKey, obtuvo: %v", err)
	}
}

func TestDecryptWithInvalidKey(t *testing.T) {
	key := "12345678901234567890123456789012"
	invalidKey := "clave-muy-corta"

	// Encriptar con clave válida
	ciphertext, _ := Encrypt("test", key)

	// Intentar desencriptar con clave inválida
	_, err := Decrypt(ciphertext, invalidKey)
	if err != ErrInvalidKey {
		t.Fatalf("Esperaba ErrInvalidKey, obtuvo: %v", err)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "99999999999999999999999999999999"

	// Encriptar con key1
	ciphertext, _ := Encrypt("test", key1)

	// Intentar desencriptar con key2
	_, err := Decrypt(ciphertext, key2)
	if err != ErrDecryptionFailed {
		t.Fatalf("Esperaba ErrDecryptionFailed, obtuvo: %v", err)
	}
}

func TestDecryptWithCorruptedData(t *testing.T) {
	key := "12345678901234567890123456789012"
	corruptedData := []byte("datos-corruptos-no-encriptados")

	_, err := Decrypt(corruptedData, key)
	if err != ErrDecryptionFailed {
		t.Fatalf("Esperaba ErrDecryptionFailed, obtuvo: %v", err)
	}
}

func TestDeriveKey(t *testing.T) {
	// Derivar clave de cualquier longitud
	derived := DeriveKey("mi-clave-secreta")

	// Verificar que tenga 32 bytes
	if len(derived) != 32 {
		t.Fatalf("Clave derivada debe tener 32 bytes, tiene: %d", len(derived))
	}

	// Verificar que sea determinística
	derived2 := DeriveKey("mi-clave-secreta")
	if derived != derived2 {
		t.Fatal("DeriveKey no es determinística")
	}

	// Verificar que claves diferentes generen resultados diferentes
	derived3 := DeriveKey("otra-clave-secreta")
	if derived == derived3 {
		t.Fatal("Claves diferentes generan el mismo resultado")
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := ""

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Error al encriptar string vacío: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Error al desencriptar string vacío: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("String vacío no se preservó. Esperado: '%s', Obtenido: '%s'", plaintext, decrypted)
	}
}

func TestEncryptDecryptLongText(t *testing.T) {
	key := "12345678901234567890123456789012"
	// Token largo simulando uno real de OAuth
	plaintext := "APP_USR-1234567890123456-123456-abcdef1234567890abcdef1234567890-1234567890"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Error al encriptar texto largo: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Error al desencriptar texto largo: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Texto largo no coincide. Esperado: %s, Obtenido: %s", plaintext, decrypted)
	}
}

func TestEncryptToBase64(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "test-token-123"

	// Encriptar a base64
	base64Cipher, err := EncryptToBase64(plaintext, key)
	if err != nil {
		t.Fatalf("Error al encriptar a base64: %v", err)
	}

	// Desencriptar desde base64
	decrypted, err := DecryptFromBase64(base64Cipher, key)
	if err != nil {
		t.Fatalf("Error al desencriptar desde base64: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Texto no coincide. Esperado: %s, Obtenido: %s", plaintext, decrypted)
	}
}
