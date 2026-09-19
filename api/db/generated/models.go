package generated

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalUrl string    `json:"original_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int64     `json:"click_count"`
}
