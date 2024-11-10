package confirmation

import "sw/internal/email"

type Data struct {
	ConfirmationToken string
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f Factory) Create(ctx email.Context[Data]) (email.Email, error) {
	subject := "Email Confirmation"
	link := "https://my-frontend/email-confirmation?token=" + ctx.Data.ConfirmationToken
	body := "Follow the link to confirm your account: " + link
	return email.Email{To: ctx.To, Subject: subject, PlainText: body}, nil
}
