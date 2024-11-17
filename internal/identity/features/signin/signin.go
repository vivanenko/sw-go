package signin

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"net/http"
	"sw/internal/cqrs"
	"sw/internal/encoding"
	"sw/internal/fluent"
	"sw/internal/identity/crypto"
	"sw/internal/logging"
)

type SignInRequest struct {
	Email    string `json:"email" validate:"required,max=320,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type SignInResponse struct {
	Token string `json:"token"`
}

func NewSignInHandler(
	logger logging.Logger,
	decoder encoding.Decoder,
	encoder encoding.Encoder,
	validate *validator.Validate,
	cmdHandler cqrs.CommandHandlerWithResponse[SignInCommand, SignInCommandResponse],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fluent.NewContextWithResponse[SignInRequest, SignInResponse](w, r).
			WithDecoder(decoder).
			WithEncoder(encoder).
			ValidatedBy(validate).
			WithHandler(func(request SignInRequest) (SignInResponse, error) {
				cmd := SignInCommand{Email: request.Email, Password: request.Password}
				cmdResponse, err := cmdHandler.Execute(cmd)
				if err != nil {
					return SignInResponse{}, err
				}
				response := SignInResponse{Token: cmdResponse.Token}
				return response, nil
			}).Handle()

		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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
