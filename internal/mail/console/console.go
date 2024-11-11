package console

import (
	"fmt"
	"sw/internal/mail"
)

type Emailer struct{}

func NewEmailer() *Emailer {
	return &Emailer{}
}

func (e Emailer) Send(email mail.Email) error {
	fmt.Println(email.PlainText)
	return nil
}
