package signup

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/fluent"
	"sw/internal/identity/crypto"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/logging"
	"sw/internal/mail"
	"time"
)

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,max=320,email,not_exist"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func NewSignUpHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	validate *validator.Validate,
	cmdHandler cqrs.CommandHandler[SignUpCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fluent.NewContext[SignUpRequest](w, r).
			WithDecoder(decoder).
			WithEncoder(encoder).
			ValidatedBy(validate).
			WithHandler(func(request SignUpRequest) error {
				cmd := SignUpCommand{Email: request.Email, Password: request.Password}
				return cmdHandler.Execute(cmd)
			}).Handle()

		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
