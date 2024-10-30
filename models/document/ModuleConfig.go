package document

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type DocumentModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewDocumentModuleConfig(DB *gorm.DB, log log.Logger) *DocumentModuleConfig {
	return &DocumentModuleConfig{DB: DB, log: log}
}

func (m *DocumentModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&Document{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}
}
