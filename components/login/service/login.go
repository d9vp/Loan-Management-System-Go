package service

import (
	"errors"
	"loanApp/models/logininfo"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type LoginService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoginService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoginService {
	return &LoginService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (l *LoginService) CreateLoginInfo(user *user.User, logininfo *logininfo.LoginInfo) error {

	user.LoginInfo = append(user.LoginInfo, logininfo)

	uow := repository.NewUnitOfWork(l.DB)
	defer uow.RollBack()

	err := l.repository.Add(uow, logininfo)
	if err != nil {
		return err
	}

	if err := l.repository.Update(uow, user); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (ls *LoginService) GetActiveLoginInfo(userID int) (*logininfo.LoginInfo, error) {
	var loginInfo logininfo.LoginInfo

	err := ls.DB.Where("user_id = ? AND logout_time IS NULL", userID).First(&loginInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &loginInfo, nil
}
