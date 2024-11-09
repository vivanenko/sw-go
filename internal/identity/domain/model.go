package domain

import (
	"net/mail"
)

type Account struct {
	Id             int64
	Email          Email
	EmailConfirmed bool
	PasswordHash   string
	PasswordSalt   string
}

func NewAccount(email Email, emailConfirmed bool, passwordHash string, passwordSalt string) *Account {
	return &Account{0, email, emailConfirmed, passwordHash, passwordSalt}
}

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	if !EmailValid(value) {
		return Email{}, &ErrInvalidEmail{}
	}
	return Email{value}, nil
}

func (e Email) String() string {
	return e.value
}

func EmailValid(email string) bool {
	ea, err := mail.ParseAddress(email)
	return err == nil && ea.Address == email
}

type ErrInvalidEmail struct{}

func (m *ErrInvalidEmail) Error() string {
	return "invalid email"
}
