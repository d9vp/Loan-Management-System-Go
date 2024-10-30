package service

import (
	"errors"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type UserService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func (u *UserService) GetUserByEmail(email string) (*user.User, error) {
	var foundUser user.User
	err := u.DB.Where("email = ?", email).First(&foundUser).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &foundUser, nil
}

func (u *UserService) AuthenticateUser(email, password string) (*user.User, error) {
	foundUser, err := u.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if foundUser.Password != password {
		return nil, errors.New("invalid password")
	}

	if !foundUser.IsActive {
		return nil, errors.New("user is inactive")
	}

	return foundUser, nil
}

func NewUserService(db *gorm.DB, repository repository.Repository, log log.Logger) *UserService {
	return &UserService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}
