package domain

type AccountRepository interface {
	Exists(email string) (bool, error)
}
