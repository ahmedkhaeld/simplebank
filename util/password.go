package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a hash of the given password using the bcrypt algorithm.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a plaintext password with a hashed password.
// Returns an error if the passwords do not match.
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
