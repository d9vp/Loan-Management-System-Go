package installation

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Installation struct {
	gorm.Model
	LoanApplicationID uint       `gorm:"not null"` //foreign key refereneces loanapplications(id)
	AmountToBePaid    float64    `gorm:"not null"`
	DueDate           time.Time  `gorm:"not null"`
	PaymentDate       *time.Time `gorm:"nullable"`
	Status            string     `gorm:"not null"`
}
