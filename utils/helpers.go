package utils

import (
	"fmt"
	"strings"
	"time"
)

// GetCurrentTime returns current time in server's location
func GetCurrentTime() time.Time {
	return time.Now()
}

// FormatRupiah formats integer amount to IDR currency string
func FormatRupiah(amount int) string {
	// Convert to positive for processing
	absAmount := amount
	if amount < 0 {
		absAmount = -amount
	}

	s := fmt.Sprintf("%d", absAmount)
	var result []string

	// Insert dots every 3 digits from the end
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		// Prepend the chunk
		if i < len(s) {
			result = append([]string{"."}, result...) // Prepend dot
		}
		result = append([]string{s[start:i]}, result...) // Prepend chunk
	}

	formatted := "Rp " + strings.Join(result, "")
	if amount < 0 {
		formatted = "- " + formatted
	}

	return formatted
}
