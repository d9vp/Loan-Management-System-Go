package loanscheme

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoanSchemeModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewLoanSchemeModuleConfig(DB *gorm.DB, log log.Logger) *LoanSchemeModuleConfig {
	return &LoanSchemeModuleConfig{DB: DB, log: log}
}

func (m *LoanSchemeModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&LoanScheme{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}
}
