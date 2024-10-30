package logininfo

import (
	"time"

	"github.com/jinzhu/gorm"
)

type LoginInfo struct {
	gorm.Model
	UserID     uint       `gorm:"not null"` //foreign key references users(ID)
	LoginTime  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LogoutTime *time.Time `gorm:"nullable"`
}
