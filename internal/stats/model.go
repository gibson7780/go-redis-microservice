package stats

import (
	"time"

	"github.com/google/uuid"
)

// Example represents a basic example entity
type Stat struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UrlID     uuid.UUID `json:"url_id" db:"url_id"`
	Count     int64     `json:"count" db:"count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
type CreateStatRequest struct {
	ID uuid.UUID
}

type GetStatRequest struct {
	ID string
}
type GetStatResponse struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UrlID     uuid.UUID `json:"url_id" db:"url_id"`
	Count     int64     `json:"count" db:"count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
