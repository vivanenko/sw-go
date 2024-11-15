package signup

import (
	"database/sql"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/identity/crypto"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/logging"
	"sw/internal/mail"
	"time"
)

type signUpRequest struct {
	Email    string `json:"email" validate:"required,max=320,email,not_exist"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func NewSignUpHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	cmdHandler cqrs.CommandHandler[SignUpCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wrapper := httpext.NewWrapper(w, r, logger, decoder, encoder)
		var request signUpRequest
		err := wrapper.Bind(&request)
		if err != nil {
			wrapper.BadRequestErr(err)
			return
		}

		cmd := SignUpCommand{Email: request.Email, Password: request.Password}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			wrapper.InternalServerError(err)
			return
		}
	}
}

type SignUpCommandHandler struct {
	db           *sql.DB
	hasher       crypto.Hasher
	emailFactory mail.Factory[confirmation.Data]
	emailer      mail.Emailer
}

type SignUpCommand struct {
	Email    string
	Password string
}

func NewSignUpCommandHandler(
	db *sql.DB,
	hasher crypto.Hasher,
	emailFactory mail.Factory[confirmation.Data],
	emailer mail.Emailer,
) *SignUpCommandHandler {
	return &SignUpCommandHandler{db: db, hasher: hasher, emailFactory: emailFactory, emailer: emailer}
}

func (h *SignUpCommandHandler) Execute(cmd SignUpCommand) error {
	passwordHash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		return err
	}

	query := "INSERT INTO account VALUES (DEFAULT, $1, $2, $3, $4) RETURNING id"
	var id int64
	err = h.db.QueryRow(query, cmd.Email, false, passwordHash, time.Now().UTC()).Scan(&id)
	if err != nil {
		return err
	}

	err = sendConfirmationToken(h.db, h.emailFactory, h.emailer, id, cmd.Email)
	return err
}
