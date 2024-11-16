package signup

import (
	"database/sql"
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"
	"sw/internal/apierr"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/fluent"
	"sw/internal/logging"
	"time"
)

const (
	ErrInvalidEmailConfirmation = "ERR_INVALID_EMAIL_CONFIRMATION"
)

type EmailConfirmationRequest struct {
	Token string `json:"token" validate:"required,max=64"`
}

func NewEmailConfirmationHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	validate *validator.Validate,
	cmdHandler cqrs.CommandHandler[EmailConfirmationCommand],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fluent.NewContext[EmailConfirmationRequest](w, r).
			WithDecoder(decoder).
			WithEncoder(encoder).
			ValidatedBy(validate).
			OnError(InvalidConfirmationError, apierr.ErrorResponse{
				Code:    ErrInvalidEmailConfirmation,
				Message: "The token is invalid or expired",
			}).
			WithHandler(func(request EmailConfirmationRequest) error {
				cmd := EmailConfirmationCommand{Token: request.Token}
				return cmdHandler.Execute(cmd)
			}).Handle()

		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
