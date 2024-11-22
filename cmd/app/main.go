package main

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"log"
	"os"
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
	"sw/internal/mail/console"
)

const (
	appConfigPath = "config/app.yml"
	migrationsSrc = "file://migrations"
)

func main() {
	connectionString := os.Getenv("SW_CONNECTION_STRING")
	secret := []byte(os.Getenv("SW_SECRET"))

	cfg, err := config.ReadConfig(appConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	logger := log.Default()
	db, err := prepareDatabase(connectionString, migrationsSrc, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}(db)
	accountRepository := postgresql.NewPgAccountRepository(db)
	validate := validator.New()
	err = validate.RegisterValidation("not_exist", validation.NewAccountNotExistValidator(accountRepository, logger))
	if err != nil {
		logger.Fatal(err)
	}
	err = validate.RegisterValidation("exists", validation.NewAccountExistsValidator(accountRepository, logger))
	if err != nil {
		logger.Fatal(err)
	}
	customValidator := validation.NewCustomValidator(validate)
	hasher := crypto.NewDefaultHasher()
	emailFactory := confirmation.NewFactory()
	emailer := console.NewEmailer()
	// SignUp
	signUpCmdHandler := signup.NewSignUpCommandHandler(db, hasher, emailFactory, emailer)
	resendEmailConfirmationCmdHandler := signup.NewResendEmailConfirmationCommandHandler(db, emailFactory, emailer)
	emailConfirmationCmdHandler := signup.NewEmailConfirmationCommandHandler(db)
	// SignIn
	signInCmdHandler := signin.NewSignInCommandHandler(cfg.JWT, secret, db, hasher)

	// Jobs
	//confirmationsCleaner := signup.NewConfirmationsCleaner(db, logger)
	//go confirmationsCleaner.Clean()

	// Web
	e := echo.New()
	e.Debug = true
	e.Validator = customValidator
	//e.HTTPErrorHandler = func(err error, c echo.Context) {
	//	logger.Println(err)
	//	c.Response().WriteHeader(http.StatusInternalServerError)
	//}
	e.Use(auth.Authentication(secret))

	e.POST("/signup", signup.NewSignUpHandler(signUpCmdHandler))
	e.POST("/resend-email-confirmation", signup.NewResendEmailConfirmationHandler(resendEmailConfirmationCmdHandler))
	e.POST("/email-confirmation", signup.NewEmailConfirmationHandler(emailConfirmationCmdHandler))
	e.POST("/signin", signin.NewSignInHandler(signInCmdHandler))
	e.GET("/me", me.NewMeHandler(), auth.Authorization())

	err = e.Start(":3000")
	if err != nil {
		logger.Fatal(err)
	}
}

func prepareDatabase(connectionString string, migrationsSrc string, logger logging.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	m, err := migrate.New(migrationsSrc, connectionString)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			logger.Println("Migrate: The database is up to date.")
		} else {
			return nil, err
		}
	}

	return db, nil
}
