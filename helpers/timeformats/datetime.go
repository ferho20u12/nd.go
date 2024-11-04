package timeformats

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateTime para manejar fechas con tiempos en formato yyyy-mm-dd hh:mm:ss
type DateTime struct {
	time.Time
}

const dateTimeLayout = "2006-01-02 15:04:05"

// UnmarshalJSON para DateTime
func (cdt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(dateTimeLayout, s)
	if err != nil {
		return err
	}
	cdt.Time = t
	return nil
}

// MarshalJSON para DateTime
func (cdt DateTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", cdt.Format(dateTimeLayout))), nil
}

// Value para DateTime para soporte de GORM
func (cdt DateTime) Value() (driver.Value, error) {
	return cdt.Format(dateTimeLayout), nil
}

// Scan para DateTime para soporte de GORM
func (cdt *DateTime) Scan(value interface{}) error {
	if value == nil {
		*cdt = DateTime{Time: time.Time{}}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan DateTime: %v", value)
	}
	*cdt = DateTime{Time: t}
	return nil
}
