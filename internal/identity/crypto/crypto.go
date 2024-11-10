package crypto

import "golang.org/x/crypto/bcrypt"

type Hasher interface {
	Hash(password string) (string, error)
	Match(hashedPassword string, currentPassword string) bool
}

type DefaultHasher struct{}

func NewDefaultHasher() *DefaultHasher {
	return &DefaultHasher{}
}

func (h DefaultHasher) Hash(password string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPasswordBytes), err
}

func (h DefaultHasher) Match(hashedPassword string, currentPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword))
	return err == nil
}
