package postgresql

import (
	"database/sql"
)

type PgAccountRepository struct {
	db *sql.DB
}

func NewPgAccountRepository(db *sql.DB) *PgAccountRepository {
	return &PgAccountRepository{db}
}

func (r *PgAccountRepository) Exists(email string) (bool, error) {
	query := "SELECT 1 FROM account WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
