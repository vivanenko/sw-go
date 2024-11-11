package signup

import (
	"database/sql"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/logging"
	"sw/internal/mail"
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
		wrapper := httpext.NewWrapper(w, r, logger, decoder, encoder)
		var request resendEmailConfirmationRequest
		err := wrapper.Bind(&request)
		if err != nil {
			return
		}

		cmd := ResendEmailConfirmationCommand{Email: request.Email}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			wrapper.InternalServerError(err)
			return
		}
	}
}

type ResendEmailConfirmationCommandHandler struct {
	db           *sql.DB
	emailFactory mail.Factory[confirmation.Data]
	emailer      mail.Emailer
}

type ResendEmailConfirmationCommand struct {
	Email string
}

func NewResendEmailConfirmationCommandHandler(
	db *sql.DB,
	emailFactory mail.Factory[confirmation.Data],
	emailer mail.Emailer,
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
		return nil
	}
	err = sendConfirmationToken(h.db, h.emailFactory, h.emailer, id, cmd.Email)
	return err
}
