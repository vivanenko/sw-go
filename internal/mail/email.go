package mail

type Email struct {
	To        string
	Subject   string
	PlainText string
	HTML      string
}

type Emailer interface {
	Send(email Email) error
}

type Context[T any] struct {
	To   string
	Data T
}

type Factory[T any] interface {
	Create(ctx Context[T]) (Email, error)
}
