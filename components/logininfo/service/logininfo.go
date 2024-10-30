package service

import (
	"loanApp/models/logininfo"
	"loanApp/repository"
	"loanApp/utils/log"

	// "time"

	"github.com/jinzhu/gorm"
)

type LoginInfoService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoginInfoService(db *gorm.DB, repo repository.Repository, logger log.Logger) *LoginInfoService {
	return &LoginInfoService{
		DB:         db,
		repository: repo,
		log:        logger,
	}
}

func (s *LoginInfoService) CreateLoginInfo(info *logininfo.LoginInfo) (*logininfo.LoginInfo, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.Add(uow, info); err != nil {
		s.log.Error("Error creating login info: ", err)
		return nil, err
	}

	uow.Commit()
	return info, nil
}

func (s *LoginInfoService) UpdateLogoutTime(userID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var info logininfo.LoginInfo
	if err := s.repository.GetByID(uow, &info, userID); err != nil {
		s.log.Error("Error fetching login info: ", err)
		return err
	}

	info.LogoutTime = nil
	if err := s.repository.Update(uow, &info); err != nil {
		s.log.Error("Error updating logout time: ", err)
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoginInfoService) GetLoginInfo(userID uint) (*logininfo.LoginInfo, error) {
	var info logininfo.LoginInfo
	if err := s.repository.GetByID(nil, &info, userID); err != nil {
		s.log.Error("Error fetching login info: ", err)
		return nil, err
	}
	return &info, nil
}
