package urls

import (
	"time"

	"github.com/google/uuid"
)

// Example represents a basic example entity
type Url struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	OriginUrl string    `json:"origin_url" db:"origin_url"`
	ShortUrl  string    `json:"short_url" db:"short_rul"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type GetUrlResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	ShortUrl  string    `json:"short_url"`
	OriginUrl string    `json:"origin_url"`
	// IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ExampleCreate represents the data needed to create a new example
type UrlCreate struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OriginUrl string    `json:"origin_url" db:"origin_url"`
	Code      string    `json:"code" db:"code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUrlRequest struct {
	// Key string `json:"key"`
	Url string `json:"url"`
}

type CreateUrlResponse struct {
	ID        uuid.UUID `json:"id"`
	OriginUrl string    `json:"origin_url"`
	ShortUrl  string    `json:"short_url"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
