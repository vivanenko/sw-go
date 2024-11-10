package main

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"sw/config"
	"sw/internal/encoding/json"
	"sw/internal/identity/crypto"
	"sw/internal/identity/features/signup"
	"sw/internal/identity/infrastructure/postgresql"
	iv "sw/internal/identity/validation"
	"sw/internal/validation"
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

	db, err := prepareDatabase(connectionString, migrationsSrc)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)
	accountRepository := postgresql.NewPgAccountRepository(db)
	validate := validator.New()
	err = validate.RegisterValidation("exists", iv.NewAccountExistsValidator(accountRepository))
	if err != nil {
		log.Fatal(err)
	}
	defaultValidator := validation.NewDefaultValidator(validate)
	encoder := json.NewEncoder()
	decoder := json.NewDecoder(defaultValidator)
	hasher := crypto.NewDefaultHasher()
	signUpCmdHandler := signup.NewCommandHandler(db, hasher)

	router := http.NewServeMux()
	router.HandleFunc("POST /signup", signup.Handler(decoder, encoder, signUpCmdHandler))
	err = http.ListenAndServe(":3000", router)
	if err != nil {
		log.Fatal(err)
	}
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

func prepareDatabase(connectionString string, migrationsSrc string) (*sql.DB, error) {
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
			log.Println("Migrate: The database is up to date.")
		} else {
			return nil, err
		}
	}

	return db, nil
}
