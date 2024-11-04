package timeformats

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Time para manejar tiempos en formato hh:mm
type Time struct {
	time.Time
}

const timeLayout = "15:04"

// UnmarshalJSON para Time
func (ct *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// MarshalJSON para Time
func (ct Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", ct.Format(timeLayout))), nil
}

// Value para Time para soporte de GORM
func (ct Time) Value() (driver.Value, error) {
	return ct.Format(timeLayout), nil
}

// Scan para Time para soporte de GORM
func (ct *Time) Scan(value interface{}) error {
	if value == nil {
		*ct = Time{Time: time.Time{}}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan Time: %v", value)
	}
	*ct = Time{Time: t}
	return nil
}
