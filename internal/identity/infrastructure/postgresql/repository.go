package postgresql

import (
	"database/sql"
	"fmt"
	"sw/internal/identity/domain"
)

type PgAccountRepository struct {
	db *sql.DB
}

func NewPgAccountRepository(db *sql.DB) *PgAccountRepository {
	return &PgAccountRepository{db}
}

func (r *PgAccountRepository) Exists(email string) (bool, error) {
	query := "SELECT 1 FROM account WHERE email = $1"
	var exists int
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *PgAccountRepository) FindByEmail(email domain.Email) (*domain.Account, error) {
	query := "SELECT id, email, email_confirmed, password_hash, password_salt FROM account WHERE email = $1"
	account := &domain.Account{}
	err := r.db.QueryRow(query, email.String()).Scan(
		&account.Id,
		&account.Email,
		&account.EmailConfirmed,
		&account.PasswordHash,
		&account.PasswordSalt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
	}
	return account, nil
}

func (r *PgAccountRepository) SaveAccount(account *domain.Account) error {
	if account.Id == 0 {
		query := "INSERT INTO account VALUES (DEFAULT, $1, $2, $3, $4) RETURNING id"
		err := r.db.QueryRow(
			query,
			account.Email.String(),
			account.EmailConfirmed,
			account.PasswordHash,
			account.PasswordSalt,
		).Scan(&account.Id)
		if err != nil {
			return err
		}
	} else {
		query := "UPDATE account SET email = $1, email_confirmed = $2, password_hash = $3, password_salt = $4 WHERE id = $5"
		res, err := r.db.Exec(
			query,
			account.Email.String(),
			account.EmailConfirmed,
			account.PasswordHash,
			account.PasswordSalt,
			account.Id,
		)
		if err != nil {
			return err
		}
		count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("account with id %d does not exist", account.Id)
		}
	}
	return nil
}
