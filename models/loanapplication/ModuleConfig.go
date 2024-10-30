package loanapplication

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoanApplicationModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewLoanApplicationModuleConfig(DB *gorm.DB, log log.Logger) *LoanApplicationModuleConfig {
	return &LoanApplicationModuleConfig{DB: DB, log: log}
}

func (m *LoanApplicationModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&LoanApplication{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}
}
