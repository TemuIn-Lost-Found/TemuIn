package utils

import (
	"time"

	"gorm.io/gorm"
)

// ExpireHighlights sets is_highlighted to false for items where highlight_expiry has passed
func ExpireHighlights(db *gorm.DB) {
	now := time.Now()
	db.Model(&struct {
		ID            int64
		IsHighlighted bool
	}{}).
		Table("core_lostitem").
		Where("is_highlighted = ? AND highlight_expiry IS NOT NULL AND highlight_expiry < ?", true, now).
		Update("is_highlighted", false)
}
