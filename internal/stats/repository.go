package stats

import (
	"database/sql"

	commonhelpers "github.com/gibson7780/go-project/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateStat(tx *sqlx.Tx, urlId uuid.UUID) error {
	CreateStat := &Stat{}

	query := `
		INSERT INTO stats (url_id)
		VALUES ($1)
		RETURNING id 
	`

	err := tx.QueryRowx(
		query,
		urlId,
	).StructScan(CreateStat)

	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetStat(id uuid.UUID) (*Stat, error) {
	var stat Stat
	err := r.db.Get(&stat, "SELECT * FROM stats WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return &stat, nil
}

func (r *repository) BatchStats(data map[string]int64) error {
	codes := make([]string, 0)
	counts := make([]int64, 0)

	for code, count := range data {
		codes = append(codes, code)
		counts = append(counts, count)
	}

	query := `
			UPDATE stats s
			SET count = v.count + s.count
			FROM urls u
			JOIN (
				SELECT unnest($1::text[]) AS code, unnest($2::bigint[]) AS count
			) v ON u.code = v.code
			WHERE s.url_id = u.id
			`

	_, err := r.db.Exec(query, pq.Array(codes), pq.Array(counts))
	return err
}
