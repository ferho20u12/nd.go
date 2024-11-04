package timeformats

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Date to handle dates in the format yyyy-mm-dd
type Date struct {
	time.Time
}

const (
	dateLayout = "2006-01-02"
)

var (
	ErrParseDate         = fmt.Errorf("error parsing date")
	ErrInvalidDateFormat = fmt.Errorf("invalid date format")
	ErrUnsupportedType   = fmt.Errorf("unsupported type for Date")
)

// StringToDate converts a date string in various formats to time.Time.
// Supported formats: "2006-01-02", Unix timestamp, "2006-01-02T15:04:05.000Z" (RFC3339)
// Returns: (*time.Time, error)
// StringToDate converts a date string or a date-time string to time.Time, extracting only the date.
// Returns: (*time.Time, error)
func StringToDate(dateStr string) (*time.Time, error) {
	var parsedDate time.Time
	var err error

	// Try to parse as a full date-time (RFC3339)
	if parsedDate, err = time.Parse(time.RFC3339, dateStr); err == nil {
		// Extract the date part
		dateOnlyStr := parsedDate.Format(dateLayout)
		if parsedDate, err = time.Parse(dateLayout, dateOnlyStr); err != nil {
			return nil, fmt.Errorf("failed to parse extracted date: %w", err)
		}
		return &parsedDate, nil
	}

	// If the input is already in date format, just parse it
	if parsedDate, err = time.Parse(dateLayout, dateStr); err == nil {
		return &parsedDate, nil
	}

	return nil, fmt.Errorf("failed to parse date: %v", err)
}

// UnmarshalJSON deserializes a date from JSON for Date.
// HTTP Status Code: 400 Bad Request if parsing fails
func (cd *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDateFormat, err)
	}
	cd.Time = t
	return nil
}

// MarshalJSON serializes the date to JSON for Date.
// HTTP Status Code: 200 OK
func (cd Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", cd.Format(dateLayout))), nil
}

// Value for Date for GORM support.
// HTTP Status Code: 200 OK
func (cd Date) Value() (driver.Value, error) {
	return cd.Format(dateLayout), nil
}

// Scan for Date for GORM support.
// HTTP Status Code: 400 Bad Request for unsupported types
func (cd *Date) Scan(value interface{}) error {
	if value == nil {
		*cd = Date{Time: time.Time{}}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*cd = Date{Time: v}
	case string:
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidDateFormat, err)
		}
		*cd = Date{Time: t}
	default:
		return fmt.Errorf("%w: %v", ErrUnsupportedType, v)
	}
	return nil
}
