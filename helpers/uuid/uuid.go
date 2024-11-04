package uuid

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UUID es una estructura que envuelve el tipo mssql.UniqueIdentifier.
type UUID [16]byte

// New crea una nueva instancia de UUID.
func New() *UUID {
	// Genera un nuevo UUID utilizando uuid.New()
	newUUID := uuid.New()
	var uuidStruct UUID
	copy(uuidStruct[:], newUUID[:])
	return &uuidStruct
}

// FromString convierte un string en un UUID.
// Si el UUID es vacío o no válido, devuelve nil.
func FromString(uuidStr string) (*UUID, error) {
	if uuidStr == "" || uuidStr == "00000000-0000-0000-0000-000000000000" {
		return nil, nil // Devuelve nil si el UUID es vacío.
	}

	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID string: %v", err)
	}

	uid := New() // Crear una nueva instancia usando New()
	copy((*uid)[:], parsedUUID[:])
	return uid, nil
}

// Scan implementa el método Scanner para convertir valores de base de datos a UUID.
func (u *UUID) Scan(value interface{}) error {
	reverse := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	switch v := value.(type) {
	case []byte:
		if len(v) != 16 {
			return errors.New("mssql: invalid UniqueIdentifier length")
		}
		var raw UUID
		copy(raw[:], v)
		reverse(raw[0:4])
		reverse(raw[4:6])
		reverse(raw[6:8])
		*u = raw
		return nil
	case string:
		if len(v) != 36 {
			return errors.New("mssql: invalid UniqueIdentifier string length")
		}
		b := []byte(v)
		for i, c := range b {
			if c == '-' {
				b = append(b[:i], b[i+1:]...)
			}
		}
		bytes, err := hex.DecodeString(string(b))
		if err != nil {
			return err
		}
		copy(u[:], bytes)
		return nil
	default:
		return fmt.Errorf("mssql: cannot convert %T to UniqueIdentifier", v)
	}
}

// Value implementa el método driver.Valuer para convertir UUID a un valor que pueda ser almacenado en la base de datos.
func (u UUID) Value() (driver.Value, error) {
	reverse := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	if u == (UUID{}) {
		return nil, nil // Devuelve NULL si el UUID es vacío.
	}

	raw := make([]byte, len(u))
	copy(raw, u[:])

	reverse(raw[0:4])
	reverse(raw[4:6])
	reverse(raw[6:8])

	return raw, nil
}

// String convierte el UUID a su representación en string.
func (u UUID) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// MarshalText convierte UniqueIdentifier a bytes correspondientes a la representación en hexadecimal.
func (u UUID) MarshalText() (text []byte, err error) {
	text = []byte(u.String())
	return
}

// UnmarshalJSON convierte una representación en string de UniqueIdentifier a bytes.
func (u *UUID) UnmarshalJSON(b []byte) error {
	// Eliminar las comillas
	input := strings.Trim(string(b), `"`)
	// Decodificar
	bytes, err := hex.DecodeString(strings.Replace(input, "-", "", -1))
	if err != nil {
		return err
	}
	// Copiar los bytes al UniqueIdentifier
	copy(u[:], bytes)
	return nil
}
