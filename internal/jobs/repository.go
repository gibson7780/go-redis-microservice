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

func (r *repository) Create(tx *sqlx.Tx, job *Job) (*Job, error) {
	query := `
		INSERT INTO jobs (type, payload)
		VALUES ($1, $2)
		RETURNING id, type, status, payload, created_at
	`
	result := &Job{}
	err := tx.QueryRowx(query, job.Type, job.Payload).StructScan(result)
	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}
	return result, nil
}

func (r *repository) GetJobs() (*[]Job, error) {
	query := `
		SELECT * 
		FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at
	`
	var jobs []Job

	err := r.db.Select(&jobs, query)
	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return &jobs, nil
}

func (r *repository) SetStatus(job *Job) error {
	query := `
	 	UPDATE jobs
    	SET status = $1, updated_at = $2
    	WHERE id = $3
	`

	_, err := r.db.Exec(query, job.Status, job.UpdatedAt, job.ID)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}

func (r *repository) CreateIdem(tx *sqlx.Tx, data *IdemData) (*IdemData, error) {

	query := `
		INSERT INTO idempotency_keys (key, status, lease, user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO NOTHING
		RETURNING id, user_id, key, status, response, lease, created_at
	`
	result := &IdemData{}
	err := tx.QueryRowx(query, data.Key, data.Status, data.Lease, data.UserID).StructScan(result)
	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}
	return result, nil
}

func (r *repository) UpdateIdem(tx *sqlx.Tx, data *IdemData) (*IdemData, error) {

	query := `
		UPDATE idempotency_keys 
		SET status = $1, response = $2
		WHERE key = $3
		RETURNING id, user_id, key, status, response, lease, created_at
	`
	result := &IdemData{}
	err := tx.QueryRowx(query, data.Key, data.Status, data.Response).StructScan(result)
	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}
	return result, nil
}
