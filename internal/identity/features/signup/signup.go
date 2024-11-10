package signup

import (
	"database/sql"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/email"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/identity/crypto"
	"sw/internal/identity/email/confirmation"
	"sw/internal/logging"
	"sw/internal/random"
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
		var request signUpRequest
		err := decoder.Decode(r.Body, &request)
		if err != nil {
			httpext.BadRequest(w, encoder, err)
			return
		}

		cmd := SignUpCommand{Email: request.Email, Password: request.Password}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			logger.Println(err)
			http.Error(w, httpext.InternalServerError, http.StatusInternalServerError)
			return
		}
	}
}

type SignUpCommandHandler struct {
	db           *sql.DB
	hasher       crypto.Hasher
	emailFactory email.Factory[confirmation.Data]
	emailer      email.Emailer
}

type SignUpCommand struct {
	Email    string
	Password string
}

func NewSignUpCommandHandler(
	db *sql.DB,
	hasher crypto.Hasher,
	emailFactory email.Factory[confirmation.Data],
	emailer email.Emailer,
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
