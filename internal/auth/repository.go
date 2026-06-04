package auth

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

func (r *repository) CreateUser(payload *User) (*User, error) {
	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id, email, password, created_at, updated_at
	`

	result := &User{}
	err := r.db.QueryRowx(
		query,
		payload.Email,
		payload.Password,
	).StructScan(result)

	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return result, nil
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)

	if err != nil {
		return nil, commonhelpers.AnalyzeDBErr(err)
	}

	return &user, nil
}
