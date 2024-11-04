package responses

import (
	"errors"

	"gorm.io/gorm"
)

const (
	TypeDB     = "DB"
	TypeBind   = "BIND"
	TypeUnknow = "UNKNOW"
)

type ErrorDetail struct {
	Error   error  `json:"-"`
	Code    int    `json:"code,omitempty"`
	Type    string `json:"type,omitempty"`
	Key     string `json:"key,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

type Error struct {
	ErrorDetail ErrorDetail `json:"error"`
	Status      int         `json:"status"`
	Code        int         `json:"-"`
}

func (e *ErrorDetail) LoadDetail() {
	if e.Details == "" {
		var errorTraducido string
		if e.Error != nil {
			switch e.Type {
			case TypeDB:
				errorTraducido = TraducirErrorDB(e.Error)
			case TypeBind:
				errorTraducido = TraducirBind(e.Error)
			default:
				errorTraducido = e.Error.Error()
			}
		} else {
			errorTraducido = "Error no definido"
		}
		e.Details = errorTraducido
	}
	if e.Type != TypeDB && e.Type != TypeBind {
		e.Type = TypeUnknow
	}
}

func TraducirErrorDB(err error) string {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "Registro no encontrado"
	} else if errors.Is(err, gorm.ErrInvalidTransaction) {
		return "Transacción inválida"
	} else if errors.Is(err, gorm.ErrNotImplemented) {
		return "Función no implementada"
	} else if errors.Is(err, gorm.ErrMissingWhereClause) {
		return "Falta cláusula WHERE"
	} else if errors.Is(err, gorm.ErrUnsupportedRelation) {
		return "Relación no soportada"
	} else if errors.Is(err, gorm.ErrPrimaryKeyRequired) {
		return "Se requiere clave primaria"
	} else if errors.Is(err, gorm.ErrModelValueRequired) {
		return "Se requiere valor del modelo"
	} else if errors.Is(err, gorm.ErrInvalidData) {
		return "Datos inválidos"
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return "Clave duplicada"
	}
	return err.Error()
}

func TraducirBind(err error) string {
	// if errGin, ok := err.(gin.Error); !ok {
	// }
	return err.Error()
}
