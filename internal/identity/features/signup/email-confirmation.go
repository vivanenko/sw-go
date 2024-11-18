package signup

import (
	"database/sql"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"sw/internal/apierr"
	"sw/internal/cqrs"
	"time"
)

const (
	ErrInvalidEmailConfirmation = "ERR_INVALID_EMAIL_CONFIRMATION"
)

type EmailConfirmationRequest struct {
	Token string `json:"token" validate:"required,max=64"`
}

func NewEmailConfirmationHandler(cmdHandler cqrs.CommandHandler[EmailConfirmationCommand]) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request EmailConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}
		err = c.Validate(request)
		if err != nil {
			return err
		}
		cmd := EmailConfirmationCommand{Token: request.Token}
		err = cmdHandler.Execute(cmd)
		if err != nil {
			if err == InvalidConfirmationError {
				err = c.JSON(http.StatusBadRequest, apierr.ErrorResponse{
					Code:    ErrInvalidEmailConfirmation,
					Message: "The token is invalid or expired",
				})
				if err != nil {
					return err
				}
			}
			return err
		}
		return nil
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
