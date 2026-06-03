package urls

import (
	"log/slog"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUrl(payload *Url) (*Url, error) {
	query := `
		INSERT INTO urls (id, origin_url, code)
		VALUES ($1, $2, $3)
		RETURNING id, origin_url, code 
	`

	result := &Url{}
	err := r.db.QueryRowx(
		query,
		payload.ID,
		payload.OriginUrl,
		payload.Code,
	).StructScan(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) GetUrl(code string) (*Url, error) {
	var response Url
	err := r.db.Get(&response, "SELECT * FROM urls WHERE code = $1", code)

	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}
	slog.Info("response", "response", response)
	return &response, nil
}
