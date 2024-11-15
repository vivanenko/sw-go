package signup

import (
	"database/sql"
	"errors"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/httpext"
	"sw/internal/logging"
	"time"
)

type emailConfirmationRequest struct {
	Token string `json:"token" validate:"required,max=64"`
}

func NewEmailConfirmationHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	cmdHandler cqrs.CommandHandler[EmailConfirmationCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wrapper := httpext.NewWrapper(w, r, logger, encoder)
		var request emailConfirmationRequest
		err := decoder.Decode(r.Body, request)
		if err != nil {
			wrapper.BadRequestErr(err)
			return
		}

		cmd := EmailConfirmationCommand{Token: request.Token}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			if err == InvalidConfirmationError {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			wrapper.InternalServerError(err)
			return
		}
	}
}

type EmailConfirmationCommandHandler struct {
	db *sql.DB
}

type EmailConfirmationCommand struct {
	Token string
}

func NewEmailConfirmationCommandHandler(db *sql.DB) *EmailConfirmationCommandHandler {
	return &EmailConfirmationCommandHandler{db: db}
}

func (h *EmailConfirmationCommandHandler) Execute(cmd EmailConfirmationCommand) error {
	exp := time.Now().AddDate(0, 0, -1)
	query := `UPDATE account a SET email_confirmed = true
				FROM email_confirmation_token t
				WHERE a.id = t.account_id AND t.value = $1 AND t.created_at > $2`
	result, err := h.db.Exec(query, cmd.Token, exp)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return InvalidConfirmationError
	}
	return nil
}

var InvalidConfirmationError = errors.New("email confirmation is invalid or expired")
