package signup

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"sw/internal/cqrs"
	"sw/internal/identity/crypto"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/mail"
	"time"
)

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,max=320,email,not_exist"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func NewSignUpHandler(cmdHandler cqrs.CommandHandler[SignUpCommand]) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request SignUpRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}
		err = c.Validate(request)
		if err != nil {
			return err
		}
		cmd := SignUpCommand{Email: request.Email, Password: request.Password}
		return cmdHandler.Execute(cmd)
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
