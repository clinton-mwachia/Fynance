package helpers

import (
	"errors"
	"regexp"
)

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

// ValidatePhoneNumber checks if the phone number starts with a country code (+XX) and is followed by 7-15 digits
func ValidatePhoneNumber(phone string) error {
	// Regular expression for international phone number format
	phoneRegex := regexp.MustCompile(`^\+\d{1,3}\d{7,15}$`)

	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format. Use +[country_code][number], e.g., +254712345678")
	}

	return nil // Phone number is valid
}
