package steganography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// EncryptData encrypts data using AES-GCM with the provided password
func EncryptData(data []byte, password string) ([]byte, error) {
	// Create a new AES cipher using the password
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData decrypts data using AES-GCM with the provided password
func DecryptData(data []byte, password string) ([]byte, error) {
	// Create a new AES cipher using the password
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check if the data is long enough
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Extract the nonce and ciphertext
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
