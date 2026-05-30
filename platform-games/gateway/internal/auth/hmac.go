package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type HMACSigner struct {
	secret string
}

func NewHMACSigner(secret string) *HMACSigner {
	return &HMACSigner{secret: secret}
}

func (h *HMACSigner) Sign(params map[string]string, timestamp int64, nonce string) string {
	sortedKeys := sortKeys(params)
	var parts []string

	for _, key := range sortedKeys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}

	paramString := strings.Join(parts, "&")
	signString := fmt.Sprintf("%s&timestamp=%d&nonce=%s", paramString, timestamp, nonce)

	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write([]byte(signString))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *HMACSigner) Verify(params map[string]string, timestamp int64, nonce, signature string) bool {
	calculatedSignature := h.Sign(params, timestamp, nonce)
	return calculatedSignature == signature
}

func sortKeys(params map[string]string) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
