package validation

import (
	"errors"
	"regexp"
)

func ValidateEmail(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return errors.New("error validating email format")
	}
	if !matched {
		return errors.New("invalid email format")
	}
	return nil
}
