package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var generateFromPassword = bcrypt.GenerateFromPassword

func HashedPassword(password string) (string, error) {
	hashedPassword, err := generateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("Failed to hash password")
	}
	return string(hashedPassword), nil
}

func CompareHashAndPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.New("Invalid email or password")
	}

	return nil
}
