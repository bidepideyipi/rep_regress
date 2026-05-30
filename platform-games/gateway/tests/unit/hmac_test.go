package unit_test

import (
	"testing"
	"time"

	"github.com/platform-games/gateway/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestHMACSigner_Sign(t *testing.T) {
	secret := "test_secret"
	signer := auth.NewHMACSigner(secret)

	params := map[string]string{
		"merchant_id": "merchant_001",
		"user_id":     "user_12345",
		"game_id":     "slot_game_v1",
	}
	timestamp := time.Now().Unix()
	nonce := "a1b2c3d4e5f6g7h8"

	signature := signer.Sign(params, timestamp, nonce)

	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 64)
}

func TestHMACSigner_Verify(t *testing.T) {
	secret := "test_secret"
	signer := auth.NewHMACSigner(secret)

	params := map[string]string{
		"merchant_id": "merchant_001",
		"user_id":     "user_12345",
		"game_id":     "slot_game_v1",
	}
	timestamp := time.Now().Unix()
	nonce := "a1b2c3d4e5f6g7h8"

	signature := signer.Sign(params, timestamp, nonce)

	isValid := signer.Verify(params, timestamp, nonce, signature)
	assert.True(t, isValid)

	isValid = signer.Verify(params, timestamp, nonce, "invalid_signature")
	assert.False(t, isValid)
}

func TestAESCipher_EncryptDecrypt(t *testing.T) {
	key := auth.GenerateAESKey()
	cipher, err := auth.NewAESCipher(key)
	assert.NoError(t, err)

	plaintext := "sensitive_data"

	encrypted, err := cipher.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := cipher.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
