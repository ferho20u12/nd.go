package facades

import (
	"framework-back/nucleo-de-diagnostico/database"

	"gorm.io/gorm"
)

func Orm() *gorm.DB {
	return database.DB.Ctx
}
