package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"
)

type KeyInfo struct {
	KeyID         string
	EncryptedKey  string
	CreatedAt     time.Time
	LastAccessedAt time.Time
	ExpiresAt     time.Time
}

type KeyMetadata struct {
	KeyType    string
	MerchantID string
	Purpose    string
}

type KeyManager struct {
	rotationInterval time.Duration
}

func NewKeyManager(rotationInterval time.Duration) *KeyManager {
	return &KeyManager{
		rotationInterval: rotationInterval,
	}
}

func (km *KeyManager) GenerateAPIKey() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate API key: %v", err))
	}
	return "api_" + hex.EncodeToString(bytes)[:32]
}

func (km *KeyManager) GenerateAPISecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate API secret: %v", err))
	}
	return hex.EncodeToString(bytes)
}

func (km *KeyManager) GenerateNonce() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate nonce: %v", err))
	}
	return hex.EncodeToString(bytes)[:16]
}

func (km *KeyManager) GenerateRSAKeyPair() (privateKey, publicKey []byte, err error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	}
	privateKey = pem.EncodeToMemory(privateKeyPEM)

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&private.PublicKey),
	}
	publicKey = pem.EncodeToMemory(publicKeyPEM)

	return privateKey, publicKey, nil
}

func (km *KeyManager) SaveRSAKeyPair(privateKeyPath, publicKeyPath string) error {
	privateKey, publicKey, err := km.GenerateRSAKeyPair()
	if err != nil {
		return err
	}

	if err := os.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := os.WriteFile(publicKeyPath, publicKey, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func (km *KeyManager) ShouldRotate(keyInfo *KeyInfo) bool {
	return time.Since(keyInfo.CreatedAt) > km.rotationInterval
}

func (km *KeyManager) RotateKey(keyID string) error {
	return errors.New("key rotation not implemented")
}

func (km *KeyManager) GenerateSessionID() string {
	timestamp := time.Now().Unix()
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("session_%d_%06d", timestamp, randomPart.Int64())
}

func (km *KeyManager) GenerateTransactionID() string {
	timestamp := time.Now().Unix()
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("txn_%d_%06d", timestamp, randomPart.Int64())
}
