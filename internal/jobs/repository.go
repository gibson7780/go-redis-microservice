package jobs

import (
	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(job *Job) (*Job, error) {
	query := `
		INSERT INTO jobs (type, payload)
		VALUES ($1, $2)
		RETURNING id, type, status, payload, created_at
	`
	result := &Job{}
	err := r.db.QueryRowx(query, job.Type, job.Payload).StructScan(result)
	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}
	return result, nil
}
