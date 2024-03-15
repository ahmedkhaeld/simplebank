package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateStringLen(value string, min int, max int) error {
	if len(value) < min || len(value) > max {
		return fmt.Errorf("must be between %d and %d", min, max)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateStringLen(value, 3, 100); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("must contain only lower cases letters, digits, or underscores")
	}
	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateStringLen(value, 3, 100); err != nil {
		return err
	}

	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}

func ValidatePassword(value string) error {
	return ValidateStringLen(value, 6, 100)
}

func ValidateEmail(value string) error {
	if err := ValidateStringLen(value, 3, 200); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is not a valid email address")
	}
	return nil
}
