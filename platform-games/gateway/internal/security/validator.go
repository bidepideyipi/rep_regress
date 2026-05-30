package security

import (
	"fmt"
	"regexp"
	"unicode"
)

var (
	userIDRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]{4,32}$`)
	merchantIDRegex = regexp.MustCompile(`^merchant_[a-zA-Z0-9]{3,28}$`)
	gameIDRegex    = regexp.MustCompile(`^[a-zA-Z0-9_-]{4,32}$`)
	apiKeyRegex    = regexp.MustCompile(`^api_[a-zA-Z0-9]{32}$`)
	nonceRegex     = regexp.MustCompile(`^[a-zA-Z0-9]{16}$`)
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateUserID(userID string) error {
	if !userIDRegex.MatchString(userID) {
		return fmt.Errorf("invalid user ID format: must be 4-32 alphanumeric characters or underscores")
	}
	return nil
}

func (v *Validator) ValidateMerchantID(merchantID string) error {
	if !merchantIDRegex.MatchString(merchantID) {
		return fmt.Errorf("invalid merchant ID format: must start with 'merchant_' followed by 3-28 alphanumeric characters")
	}
	return nil
}

func (v *Validator) ValidateGameID(gameID string) error {
	if !gameIDRegex.MatchString(gameID) {
		return fmt.Errorf("invalid game ID format: must be 4-32 alphanumeric characters, hyphens or underscores")
	}
	return nil
}

func (v *Validator) ValidateAPIKey(apiKey string) error {
	if !apiKeyRegex.MatchString(apiKey) {
		return fmt.Errorf("invalid API key format")
	}
	return nil
}

func (v *Validator) ValidateNonce(nonce string) error {
	if !nonceRegex.MatchString(nonce) {
		return fmt.Errorf("invalid nonce format: must be 16 alphanumeric characters")
	}
	return nil
}

func (v *Validator) ValidateBetAmount(amount float64) error {
	if amount < 0.1 || amount > 1000 {
		return fmt.Errorf("invalid bet amount: must be between 0.1 and 1000")
	}

	if !isMultipleOf(amount, 0.1) {
		return fmt.Errorf("invalid bet amount: must be a multiple of 0.1")
	}

	return nil
}

func (v *Validator) ValidateTimestamp(timestamp int64, maxAgeSeconds int64) error {
	currentTime := getCurrentTimestamp()
	diff := currentTime - timestamp

	if diff < 0 {
		diff = -diff
	}

	if diff > maxAgeSeconds {
		return fmt.Errorf("timestamp too old or in future: difference %d seconds", diff)
	}

	return nil
}

func (v *Validator) SanitizeString(input string) string {
	var result []rune
	for _, r := range input {
		if r == '<' || r == '>' || r == '"' || r == '\'' || r == '&' {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

func (v *Validator) ValidateString(input string, maxLength int) error {
	if len(input) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	if len(input) > maxLength {
		return fmt.Errorf("input too long: maximum %d characters", maxLength)
	}

	if containsControlChars(input) {
		return fmt.Errorf("input contains invalid control characters")
	}

	return nil
}

func isMultipleOf(value, multiple float64) bool {
	return int(value/multiple)*multiple == int(value)
}

func containsControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}

func getCurrentTimestamp() int64 {
	return 0
}
