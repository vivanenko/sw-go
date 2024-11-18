package signin

import (
	"database/sql"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"sw/internal/apierr"
	"sw/internal/cqrs"
	"sw/internal/identity/crypto"
	"sw/internal/random"
	"time"
)

const (
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
)

type SignInRequest struct {
	Email    string `json:"email" validate:"required,max=320,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type SignInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
			if err == InvalidCredentialsError {
				return c.JSON(http.StatusBadRequest, apierr.ErrorResponse{
					Code:    ErrInvalidCredentials,
					Message: "Credentials are invalid"},
				)
			}
			return err
		}
		response := SignInResponse{AccessToken: cmdResponse.AccessToken, RefreshToken: cmdResponse.RefreshToken}
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
	AccessToken  string
	RefreshToken string
}

func NewSignInCommandHandler(
	db *sql.DB,
	hasher crypto.Hasher,
) *SignInCommandHandler {
	return &SignInCommandHandler{db: db, hasher: hasher}
}

func (h *SignInCommandHandler) Execute(cmd SignInCommand) (SignInCommandResponse, error) {
	query := "SELECT id, email, email_confirmed, password_hash FROM account WHERE email = $1"
	var id string
	var email string
	var emailConfirmed bool
	var passwordHash string
	err := h.db.QueryRow(query, cmd.Email).Scan(&id, &email, &emailConfirmed, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return SignInCommandResponse{}, InvalidCredentialsError
		}
		return SignInCommandResponse{}, err
	}
	if h.hasher.Match(passwordHash, cmd.Password) {
		// todo: extract to config key and ttls
		secretKey := []byte("TODO")
		claims := jwt.MapClaims{
			"sub":   id,
			"exp":   time.Now().Add(time.Minute * 30).Unix(),
			"email": email,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessToken, err := token.SignedString(secretKey)
		if err != nil {
			return SignInCommandResponse{}, err
		}
		refreshToken := random.String(64)
		expiresAt := time.Now().UTC().AddDate(0, 0, 90)
		query = "INSERT INTO refresh_token VALUES (DEFAULT, $1, $2, $3)"
		_, err = h.db.Exec(query, refreshToken, expiresAt, id)
		if err != nil {
			return SignInCommandResponse{}, err
		}
		return SignInCommandResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
	}
	return SignInCommandResponse{}, InvalidCredentialsError
}

var InvalidCredentialsError = errors.New("invalid credentials")
