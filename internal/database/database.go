package database

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"sw/internal/logging"
)

func New(connectionString string, migrationsSrc string, logger logging.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	m, err := migrate.New(migrationsSrc, connectionString)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			logger.Println("Migrate: The database is up to date.")
		} else {
			return nil, err
		}
	}

	return db, nil
}
