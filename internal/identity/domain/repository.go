package domain

type AccountRepository interface {
	Exists(email string) bool
	//FindByEmail(email Email) (*Account, error)
}
