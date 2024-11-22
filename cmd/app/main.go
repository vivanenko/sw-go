package main

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"log"
	"os"
	"sw/config"
	"sw/internal/auth"
	"sw/internal/database"
	"sw/internal/identity"
	"sw/internal/mail/console"
	"sw/internal/validation"
)

const (
	appConfigPath = "config/app.yml"
	migrationsSrc = "file://migrations"
)

func main() {
	port := os.Getenv("SW_PORT")
	connectionString := os.Getenv("SW_CONNECTION_STRING")
	secret := []byte(os.Getenv("SW_SECRET"))

	logger := log.Default()
	cfg, err := config.ReadConfig(appConfigPath)
	if err != nil {
		logger.Fatal(err)
	}
	db, err := database.New(connectionString, migrationsSrc, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}(db)
	validate := validator.New()
	emailer := console.NewEmailer()

	e := echo.New()
	e.Debug = true
	e.Validator = validation.NewCustomValidator(validate)
	//e.HTTPErrorHandler = func(err error, c echo.Context) {
	//	logger.Println(err)
	//	c.Response().WriteHeader(http.StatusInternalServerError)
	//}
	e.Use(auth.Authentication(secret))

	err = identity.Initialize(e, logger, validate, cfg, secret, db, emailer)
	if err != nil {
		logger.Fatal(err)
	}

	err = e.Start(":" + port)
	if err != nil {
		logger.Fatal(err)
	}
}
