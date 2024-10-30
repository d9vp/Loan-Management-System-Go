package logininfo

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoginInfoModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewLoginInfoModuleConfig(DB *gorm.DB, log log.Logger) *LoginInfoModuleConfig {
	return &LoginInfoModuleConfig{DB: DB, log: log}
}

func (m *LoginInfoModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&LoginInfo{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}
}
