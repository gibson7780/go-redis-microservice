package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobsCreateResponse struct {
	ID uuid.UUID `json:"id"`
	// Type    string          `json:"type"`
	// Payload json.RawMessage `json:"payload"`
	// Status  string          `json:"status"`
}

type JobsCreateRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Job struct {
	ID        uuid.UUID       `db:"id"`
	Type      string          `db:"type"`
	Payload   json.RawMessage `db:"payload"`
	Status    string          `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

type IdemData struct {
	ID        uuid.UUID       `db:"id"`
	UserID    uuid.UUID       `db:"user_id"`
	Key       string          `db:"key"`
	Status    string          `db:"status"`
	Lease     time.Time       `db:"lease"`
	Response  json.RawMessage `db:"response"`
	CreatedAt time.Time       `db:"created_at"`
}
