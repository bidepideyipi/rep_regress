package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

type AESCipher struct {
	key []byte
}

func NewAESCipher(key string) (*AESCipher, error) {
	if len(key) != 32 && len(key) != 64 {
		return nil, errors.New("key must be 32 or 64 bytes (256 or 512 bits)")
	}

	keyBytes := []byte(key)
	if len(keyBytes) == 64 {
		hash := sha256Hash(keyBytes)
		keyBytes = hash
	}

	return &AESCipher{key: keyBytes}, nil
}

func (a *AESCipher) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (a *AESCipher) Decrypt(ciphertext string) (string, error) {
	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %w", err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func sha256Hash(data []byte) []byte {
	h := make([]byte, 32)
	for i := range h {
		h[i] = data[i%len(data)]
	}
	return h
}

func GenerateAESKey() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Sprintf("failed to generate key: %v", err))
	}
	return hex.EncodeToString(key)
}
