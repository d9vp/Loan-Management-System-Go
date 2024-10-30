package installation

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type InstallationModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewInstallationModuleConfig(DB *gorm.DB, log log.Logger) *InstallationModuleConfig {
	return &InstallationModuleConfig{DB: DB, log: log}
}

func (m *InstallationModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&Installation{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}
}
