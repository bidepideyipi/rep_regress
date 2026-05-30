package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	keyType := flag.String("type", "rsa", "Key type to generate: rsa, api, aes, nonce")
	outputDir := flag.String("output", "./keys", "Output directory for generated keys")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	switch *keyType {
	case "rsa":
		generateRSAKeys(*outputDir)
	case "api":
		generateAPIKeys()
	case "aes":
		generateAESKey()
	case "nonce":
		generateNonce()
	default:
		fmt.Println("Unknown key type. Use: rsa, api, aes, or nonce")
		os.Exit(1)
	}
}

func generateRSAKeys(outputDir string) {
	fmt.Println("Generating RSA-2048 key pair...")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyFile, err := os.Create(outputDir + "/private.pem")
	if err != nil {
		log.Fatalf("Failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyFile, err := os.Create(outputDir + "/public.pem")
	if err != nil {
		log.Fatalf("Failed to create public key file: %v", err)
	}
	defer publicKeyFile.Close()

	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		log.Fatalf("Failed to write public key: %v", err)
	}

	fmt.Println("RSA keys generated:")
	fmt.Println("  Private key:", outputDir+"/private.pem")
	fmt.Println("  Public key: ", outputDir+"/public.pem")
}

func generateAPIKeys() {
	apiKeyBytes := make([]byte, 16)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}
	apiKey := "api_" + hex.EncodeToString(apiKeyBytes)[:32]

	apiSecretBytes := make([]byte, 32)
	if _, err := rand.Read(apiSecretBytes); err != nil {
		log.Fatalf("Failed to generate API secret: %v", err)
	}
	apiSecret := hex.EncodeToString(apiSecretBytes)

	fmt.Println("API credentials generated:")
	fmt.Println("  API Key:   ", apiKey)
	fmt.Println("  API Secret:", apiSecret)
}

func generateAESKey() {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		log.Fatalf("Failed to generate AES key: %v", err)
	}
	aesKey := hex.EncodeToString(keyBytes)

	fmt.Println("AES-256 key generated:")
	fmt.Println("  Key:", aesKey)
}

func generateNonce() {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate nonce: %v", err)
	}
	nonce := hex.EncodeToString(bytes)[:16]

	fmt.Println("Nonce generated:")
	fmt.Println("  Nonce:", nonce)
}

func generateSessionID() string {
	timestamp := time.Now().Unix()
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("session_%d_%06d", timestamp, randomPart.Int64())
}

func generateTransactionID() string {
	timestamp := time.Now().Unix()
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("txn_%d_%06d", timestamp, randomPart.Int64())
}
