package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrInvalidKey se lanza cuando la clave de encriptación es inválida
	ErrInvalidKey = errors.New("invalid encryption key")
	// ErrDecryptionFailed se lanza cuando falla la desencriptación
	ErrDecryptionFailed = errors.New("decryption failed")
)

// Encrypt encripta un texto plano usando AES-GCM
// key debe ser una cadena de 32 bytes (256 bits)
func Encrypt(plaintext string, key string) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	// Crear el bloque AES
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// Crear GCM (Galois/Counter Mode)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Crear un nonce aleatorio
	// El nonce debe ser único para cada encriptación con la misma clave
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encriptar y autenticar
	// El nonce se añade al principio del texto cifrado
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// Decrypt desencripta un texto cifrado usando AES-GCM
// key debe ser la misma clave usada para encriptar
func Decrypt(ciphertext []byte, key string) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}

	// Validar que el texto cifrado tenga al menos el tamaño del nonce
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptionFailed
	}

	// Extraer el nonce del principio del texto cifrado
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Desencriptar y verificar
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// DeriveKey deriva una clave de 32 bytes a partir de una clave de cualquier longitud
// Útil para convertir claves configuradas por el usuario en claves válidas de 256 bits
func DeriveKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return string(hash[:])
}

// EncryptToBase64 encripta y retorna el resultado en base64 (útil para debugging)
func EncryptToBase64(plaintext string, key string) (string, error) {
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptFromBase64 desencripta desde base64 (útil para debugging)
func DecryptFromBase64(base64Ciphertext string, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		return "", err
	}
	return Decrypt(ciphertext, key)
}
