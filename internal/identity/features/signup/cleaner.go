package signup

import (
	"database/sql"
	"sw/internal/logging"
	"time"
)

type ConfirmationsCleaner struct {
	db     *sql.DB
	logger logging.Logger
}

func NewConfirmationsCleaner(db *sql.DB, logger logging.Logger) *ConfirmationsCleaner {
	return &ConfirmationsCleaner{db: db, logger: logger}
}

func (c *ConfirmationsCleaner) Clean() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.cleanupDatabase()
			if err != nil {
				c.logger.Println("An error occurred during confirmations cleaning:", err)
			}
		}
	}
}

func (c *ConfirmationsCleaner) cleanupDatabase() error {
	exp := time.Now().AddDate(0, 0, -1)
	query := "DELETE FROM email_confirmation_token WHERE created_at < $1"
	result, err := c.db.Exec(query, exp)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	c.logger.Println("Expired tokens deleted: %d", count)
	return nil
}
