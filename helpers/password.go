package helpers

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(passwordHash), err
}

func VerifyPassword(hashPasword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashPasword), []byte(password))
	return err
}
