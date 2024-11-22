package identity

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"sw/config"
	"sw/internal/auth"
	"sw/internal/identity/crypto"
	"sw/internal/identity/features/me"
	"sw/internal/identity/features/signin"
	"sw/internal/identity/features/signup"
	"sw/internal/identity/infrastructure/postgresql"
	"sw/internal/identity/mail/confirmation"
	"sw/internal/identity/validation"
	"sw/internal/logging"
	"sw/internal/mail"
)

func Initialize(
	e *echo.Echo,
	logger logging.Logger,
	validate *validator.Validate,
	cfg config.Config,
	secret []byte,
	db *sql.DB,
	emailer mail.Emailer,
) error {
	accountRepository := postgresql.NewPgAccountRepository(db)

	err := validate.RegisterValidation("not_exist", validation.NewAccountNotExistValidator(accountRepository, logger))
	if err != nil {
		return err
	}
	err = validate.RegisterValidation("exists", validation.NewAccountExistsValidator(accountRepository, logger))
	if err != nil {
		return err
	}

	hasher := crypto.NewDefaultHasher()
	emailFactory := confirmation.NewFactory()

	// SignUp
	signUpCmdHandler := signup.NewSignUpCommandHandler(db, hasher, emailFactory, emailer)
	resendEmailConfirmationCmdHandler := signup.NewResendEmailConfirmationCommandHandler(db, emailFactory, emailer)
	emailConfirmationCmdHandler := signup.NewEmailConfirmationCommandHandler(db)
	// SignIn
	signInCmdHandler := signin.NewSignInCommandHandler(cfg.JWT, secret, db, hasher)

	e.POST("/signup", signup.NewSignUpHandler(signUpCmdHandler))
	e.POST("/resend-email-confirmation", signup.NewResendEmailConfirmationHandler(resendEmailConfirmationCmdHandler))
	e.POST("/email-confirmation", signup.NewEmailConfirmationHandler(emailConfirmationCmdHandler))
	e.POST("/signin", signin.NewSignInHandler(signInCmdHandler))
	e.GET("/me", me.NewMeHandler(), auth.Authorization())

	// Jobs
	confirmationsCleaner := signup.NewConfirmationsCleaner(db, logger)
	go confirmationsCleaner.Clean()

	return nil
}
