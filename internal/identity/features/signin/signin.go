package signin

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/identity/crypto"
)

type SignInRequest struct {
	Email    string `json:"email" validate:"required,max=320,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type SignInResponse struct {
	Token string `json:"token"`
}

func NewSignInHandler(
	cmdHandler cqrs.CommandHandlerWithResponse[SignInCommand, SignInCommandResponse],
) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request SignInRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}
		err = c.Validate(request)
		if err != nil {
			return err
		}
		cmd := SignInCommand{Email: request.Email, Password: request.Password}
		cmdResponse, err := cmdHandler.Execute(cmd)
		if err != nil {
			return err
		}
		response := SignInResponse{Token: cmdResponse.Token}
		return c.JSON(http.StatusOK, response)
	}
}

type SignInCommandHandler struct {
	db     *sql.DB
	hasher crypto.Hasher
}

type SignInCommand struct {
	Email    string
	Password string
}

type SignInCommandResponse struct {
	Token string
}

func NewSignInCommandHandler(
	db *sql.DB,
	hasher crypto.Hasher,
) *SignInCommandHandler {
	return &SignInCommandHandler{db: db, hasher: hasher}
}

func (h *SignInCommandHandler) Execute(cmd SignInCommand) (SignInCommandResponse, error) {
	return SignInCommandResponse{Token: "qwe"}, nil
}
