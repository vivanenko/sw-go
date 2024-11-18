package main

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"sw/config"
	"sw/internal/identity/crypto"
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
	connectionString := os.Getenv("CONNECTION_STRING")

	//cfg, err := getConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}
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
	customValidator := &CustomValidator{validator: validate}
	//defaultValidator := validation.NewDefaultValidator(validate)
	hasher := crypto.NewDefaultHasher()
	emailFactory := confirmation.NewFactory()
	emailer := console.NewEmailer()
	// SignUp
	signUpCmdHandler := signup.NewSignUpCommandHandler(db, hasher, emailFactory, emailer)
	resendEmailConfirmationCmdHandler := signup.NewResendEmailConfirmationCommandHandler(db, emailFactory, emailer)
	emailConfirmationCmdHandler := signup.NewEmailConfirmationCommandHandler(db)
	// SignIn
	signInCmdHandler := signin.NewSignInCommandHandler(db, hasher)

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
	e.POST("/signup", signup.NewSignUpHandler(signUpCmdHandler))
	e.POST("/resend-email-confirmation", signup.NewResendEmailConfirmationHandler(resendEmailConfirmationCmdHandler))
	e.POST("/email-confirmation", signup.NewEmailConfirmationHandler(emailConfirmationCmdHandler))
	e.POST("/signin", signin.NewSignInHandler(signInCmdHandler))

	err = e.Start(":3000")
	if err != nil {
		logger.Fatal(err)
	}
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err)
		//return err
	}
	return nil
}

func getConfig() (*config.Config, error) {
	file, err := os.Open(appConfigPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	cfg := &config.Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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
