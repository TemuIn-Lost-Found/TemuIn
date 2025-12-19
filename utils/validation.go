package utils

import (
	"strings"
	"unicode"
)

// ValidatePassword checks if password meets requirements:
// - Minimum 8 characters
// - Contains both uppercase and lowercase letters
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password minimal 8 karakter"
	}

	hasUpper := false
	hasLower := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
	}

	if !hasUpper || !hasLower {
		return false, "Password harus mengandung huruf besar dan kecil"
	}

	return true, ""
}

// ValidateNotEmpty checks if a field is not empty
func ValidateNotEmpty(value string, fieldName string) (bool, string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false, fieldName + " wajib diisi"
	}
	return true, ""
}
