package service

import (
	"errors"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"time"

	"github.com/jinzhu/gorm"
)

type LogoutService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLogoutService(db *gorm.DB, repository repository.Repository, log log.Logger) *LogoutService {
	return &LogoutService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (l *LogoutService) UpdateLoginInfo(user *user.User) error {
	uow := repository.NewUnitOfWork(l.DB)
	defer uow.RollBack()

	user, err := l.repository.GetUserWithLoginInfo(uow, user.ID)
	if err != nil {
		return err
	}

	if len(user.LoginInfo) == 0 {
		return errors.New("no login info found for user")
	}

	now := time.Now()
	logininfo := user.LoginInfo[len(user.LoginInfo)-1]

	if logininfo.ID == 0 {
		return errors.New("invalid login info ID")
	}

	logininfo.LogoutTime = &now

	err = l.repository.Update(uow, logininfo)
	if err != nil {
		return err
	}

	if err := l.repository.Update(uow, user); err != nil {
		return err
	}

	uow.Commit()
	return nil
}
