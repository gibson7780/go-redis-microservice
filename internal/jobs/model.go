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
	ID         uuid.UUID       `db:"id"`
	Type       string          `db:"type"`
	Payload    json.RawMessage `db:"payload"`
	Status     string          `db:"status"`
	Created_at time.Time       `db:"created_at"`
}
