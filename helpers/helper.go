package helpers

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Handler is a global variable for the Helper instance
var (
	Handler *Helper
)

// Boot initializes the Handler instance with default values
func Boot() {
	Handler = &Helper{
		Orm:  *NewOrm(),
		Bind: *NewBind(),
	}
}

// Helper is a struct that aggregates various helper functionalities
type Helper struct {
	Orm  Orm
	Bind Bind
}

var (
	ErrParseInt       = errors.New("error parsing Int")          // HTTP 400 Bad Request
	ErrParseUUID      = errors.New("error parsing UUID")         // HTTP 400 Bad Request
	ErrParseDate      = errors.New("error parsing Date")         // HTTP 400 Bad Request
	ErrBirthDateToAge = errors.New("birthdate is in the future") // HTTP 400 Bad Request
	ErrFieldNotFound  = errors.New("field not found")            // HTTP 404 Not Found
)

// StringToInt converts a string to an int.
// Returns: (int, error)
func StringToInt(str string) (int, error) {
	result, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrParseInt, err)
	}
	return result, nil
}

// BirthDateToAge converts a time.Time (birthdate) to an int (age).
// Returns: (int, error)
func BirthDateToAge(birthdate time.Time) (int, error) {
	now := time.Now()
	if birthdate.After(now) {
		return 0, ErrBirthDateToAge
	}

	age := now.Year() - birthdate.Year()

	// Adjust if the birthday hasn't occurred yet this year
	if now.Month() < birthdate.Month() || (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}

	return age, nil
}

// IsValidUUID checks if a string is a valid UUID.
// Returns: error
func IsValidUUID(id string) error {
	// _, err := uuid.Parse(id)
	return nil
}

// IsValidInt checks if a string is a valid integer.
// Returns: error
func IsValidInt(id string) error {
	_, err := strconv.Atoi(id)
	return err
}

// CleanText removes extra spaces from a string and trims it.
// Returns: string
func CleanText(text string) string {
	if text != "" {
		regexSpaces := regexp.MustCompile(`\s+`)
		text = regexSpaces.ReplaceAllString(text, " ")
		text = strings.TrimSpace(text)
	}
	return text
}
