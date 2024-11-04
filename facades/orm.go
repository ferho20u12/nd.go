package facades

import (
	"github.com/ferho20u12/nd.go/database"

	"gorm.io/gorm"
)

func Orm() *gorm.DB {
	return database.DB.Ctx
}
