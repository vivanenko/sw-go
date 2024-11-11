package signup

import (
	"database/sql"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/mail"
	"sw/internal/random"
	"time"
)

func sendConfirmationToken(
	db *sql.DB,
	emailFactory mail.Factory[confirmation.Data],
	emailer mail.Emailer,
	id int64,
	email string,
) error {
	token := random.String(64)
	query := "INSERT INTO email_confirmation_token VALUES (DEFAULT, $1, $2, $3)"
	_, err := db.Exec(query, token, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	ctx := mail.Context[confirmation.Data]{To: email, Data: confirmation.Data{ConfirmationToken: token}}
	e, err := emailFactory.Create(ctx)
	if err != nil {
		return err
	}
	err = emailer.Send(e)
	return err
}
