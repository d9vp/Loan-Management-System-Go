package document

import "github.com/jinzhu/gorm"

type Document struct {
	gorm.Model
	LoanApplicationID uint    `gorm:"not null"` //int->uint //foreign key refereneces loanapplications(id)
	DocumentType      string `gorm:"not null"`
	URL               string `gorm:"not null"`
}
