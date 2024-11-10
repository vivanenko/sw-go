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
	"time"
)

type request struct {
	Email    string `json:"email" validate:"required,max=320,email,exists"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func NewHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	cmdHandler cqrs.CommandHandler[Command],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request request
		err := decoder.Decode(r.Body, &request)
		if err != nil {
			httpext.BadRequest(w, encoder, err)
			return
		}

		cmd := Command{Email: request.Email, Password: request.Password}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			logger.Println(err)
			http.Error(w, httpext.InternalServerError, http.StatusInternalServerError)
			return
		}
	}
}

type CommandHandler struct {
	db           *sql.DB
	hasher       crypto.Hasher
	emailFactory email.Factory[confirmation.Data]
	emailer      email.Emailer
}

func NewCommandHandler(
	db *sql.DB,
	hasher crypto.Hasher,
	emailFactory email.Factory[confirmation.Data],
	emailer email.Emailer,
) *CommandHandler {
	return &CommandHandler{db: db, hasher: hasher, emailFactory: emailFactory, emailer: emailer}
}

func (h *CommandHandler) Execute(cmd Command) error {
	passwordHash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		return err
	}

	query := "INSERT INTO account VALUES (DEFAULT, $1, $2, $3, $4) RETURNING id"
	_, err = h.db.Exec(
		query,
		cmd.Email,
		false,
		passwordHash,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	// todo: generate token & use transaction for token insert
	token := "some-test-token"
	ctx := email.Context[confirmation.Data]{To: cmd.Email, Data: confirmation.Data{ConfirmationToken: token}}
	e, err := h.emailFactory.Create(ctx)
	if err != nil {
		return err
	}
	err = h.emailer.Send(e)
	return err
}

type Command struct {
	Email    string
	Password string
}
