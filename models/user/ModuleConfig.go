package user

import (
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type UserModuleConfig struct {
	DB  *gorm.DB
	log log.Logger
}

func NewUserModuleConfig(DB *gorm.DB, log log.Logger) *UserModuleConfig {
	return &UserModuleConfig{DB: DB, log: log}
}

//------------??????????

func (m *UserModuleConfig) TableMigration() {
	var models []interface{} = []interface{}{
		&User{},
		&Admin{},
		&LoanOfficer{},
		&Customer{},
	}

	for _, model := range models {
		err := m.DB.AutoMigrate(model).Error
		if err != nil {
			m.log.Error(err.Error())
		}
	}

}
