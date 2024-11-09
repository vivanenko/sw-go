package signup

import (
	"database/sql"
	"log"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/httpext"
)

type request struct {
	Email    string `json:"email" validate:"required,email,exists"`
	Password string `json:"password" validate:"required,min=8,max=128"`
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
	db *sql.DB
}

func NewCommandHandler(db *sql.DB) *CommandHandler {
	return &CommandHandler{db: db}
}

func (h *CommandHandler) Execute(cmd Command) error {
	passwordHash := ""
	passwordSalt := ""

	query := "INSERT INTO account VALUES (DEFAULT, $1, $2, $3, $4) RETURNING id"
	_, err := h.db.Exec(
		query,
		cmd.Email,
		false,
		passwordHash,
		passwordSalt,
	)
	return err
}

type Command struct {
	Email    string
	Password string
}
