package signup

import (
	"database/sql"
	"log"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/identity/crypto"
	"time"
)

type request struct {
	Email    string `json:"email" validate:"required,max=320,email,exists"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func Handler(
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
			log.Println(err)
			http.Error(w, httpext.InternalServerError, http.StatusInternalServerError)
			return
		}
	}
}

type CommandHandler struct {
	db     *sql.DB
	hasher crypto.Hasher
}

func NewCommandHandler(db *sql.DB, hasher crypto.Hasher) *CommandHandler {
	return &CommandHandler{db: db, hasher: hasher}
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
	return err
}

type Command struct {
	Email    string
	Password string
}
