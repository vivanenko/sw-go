package signup

import (
	"database/sql"
	"errors"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/email"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/identity/email/confirmation"
	"sw/internal/logging"
	"sw/internal/random"
	"time"
)

type resendEmailConfirmationRequest struct {
	Email string `json:"email" validate:"required,max=320,email,exists"`
}

func NewResendEmailConfirmationHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	cmdHandler cqrs.CommandHandler[ResendEmailConfirmationCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request resendEmailConfirmationRequest
		err := decoder.Decode(r.Body, &request)
		if err != nil {
			httpext.BadRequest(w, encoder, err)
			return
		}

		cmd := ResendEmailConfirmationCommand{Email: request.Email}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			logger.Println(err)
			http.Error(w, httpext.InternalServerError, http.StatusInternalServerError)
			return
		}
	}
}

type ResendEmailConfirmationCommandHandler struct {
	db           *sql.DB
	emailFactory email.Factory[confirmation.Data]
	emailer      email.Emailer
}

type ResendEmailConfirmationCommand struct {
	Email string
}

func NewResendEmailConfirmationCommandHandler(
	db *sql.DB,
	emailFactory email.Factory[confirmation.Data],
	emailer email.Emailer,
) *ResendEmailConfirmationCommandHandler {
	return &ResendEmailConfirmationCommandHandler{db: db, emailFactory: emailFactory, emailer: emailer}
}

func (h *ResendEmailConfirmationCommandHandler) Execute(cmd ResendEmailConfirmationCommand) error {
	query := "SELECT id, email_confirmed FROM account WHERE email = $1"
	var id int64
	var emailConfirmed bool
	err := h.db.QueryRow(query, cmd.Email).Scan(&id, &emailConfirmed)
	if err != nil {
		return err
	}
	if emailConfirmed {
		return errors.New("email is already confirmed")
	}

	token := random.String(64)
	query = "INSERT INTO email_confirmation_token VALUES (DEFAULT, $1, $2, $3)"
	_, err = h.db.Exec(query, token, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	ctx := email.Context[confirmation.Data]{To: cmd.Email, Data: confirmation.Data{ConfirmationToken: token}}
	e, err := h.emailFactory.Create(ctx)
	if err != nil {
		return err
	}
	err = h.emailer.Send(e)
	return err
}
