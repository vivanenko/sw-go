package signup

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
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
	validate *validator.Validate,
	cmdHandler cqrs.CommandHandler[ResendEmailConfirmationCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request resendEmailConfirmationRequest
		err := decoder.Decode(r.Body, &request)
		if err != nil {
			// todo: message: Invalid JSON
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = validate.Struct(request)
		if err != nil {
			// todo: message: details
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cmd := ResendEmailConfirmationCommand{Email: request.Email}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
