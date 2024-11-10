package console

import (
	"fmt"
	"sw/internal/email"
)

type Emailer struct{}

func NewEmailer() *Emailer {
	return &Emailer{}
}

func (e Emailer) Send(email email.Email) error {
	fmt.Println(email.PlainText)
	return nil
}
