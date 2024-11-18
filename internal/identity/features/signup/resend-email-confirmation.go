package signup

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"sw/internal/cqrs"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/mail"
)

type ResendEmailConfirmationRequest struct {
	Email string `json:"email" validate:"required,max=320,email,exists"`
}

func NewResendEmailConfirmationHandler(
	cmdHandler cqrs.CommandHandler[ResendEmailConfirmationCommand],
) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request ResendEmailConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}
		err = c.Validate(request)
		if err != nil {
			return err
		}
		cmd := ResendEmailConfirmationCommand{Email: request.Email}
		return cmdHandler.Execute(cmd)
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
