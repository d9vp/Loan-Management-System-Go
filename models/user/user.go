package user

import (
	"loanApp/models/loanapplication"
	"loanApp/models/logininfo"

	"github.com/jinzhu/gorm"
)

type Role string

type User struct {
	gorm.Model
	Name      string                 `gorm:"not null"`
	Email     string                 `gorm:"unique;not null"`
	Password  string                 `gorm:"not null"`
	IsActive  bool                   `gorm:"default:true"`
	Role      Role                   `gorm:"not null"`
	LoginInfo []*logininfo.LoginInfo `gorm:"foreignKey:UserID"`
}

type Admin struct {
	User
	LoanOfficers []*LoanOfficer `gorm:"foreignKey:CreatedByAdminID"`
}

type LoanOfficer struct {
	User
	CreatedByAdminID uint //foreign key references admin(id)
	// UpdatedBy        []*Admin
	AssignedLoans    []*loanapplication.LoanApplication `gorm:"foreignKey:LoanOfficerID"`
}

type Customer struct {
	User
	LoanApplications []*loanapplication.LoanApplication `gorm:"foreignKey:CustomerID"`
}
