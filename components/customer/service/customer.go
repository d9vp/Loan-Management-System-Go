package service

import (
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"

	"github.com/jinzhu/gorm"
)

type CustomerService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewCustomerService(db *gorm.DB, repo repository.Repository, logger log.Logger) *CustomerService {
	return &CustomerService{
		DB:         db,
		repository: repo,
		log:        logger,
	}
}

func (s *CustomerService) CreateCustomer(customer *user.Customer) (*user.Customer, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	err := s.repository.Add(uow, &customer.User)
	if err != nil {
		return nil, err
	}

	customer.ID = customer.User.ID
	err = s.repository.Add(uow, customer)
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return customer, nil
}

func (s *CustomerService) UpdateCustomer(customer *user.Customer) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.Update(uow, customer); err != nil {
		s.log.Error("Error updating customer: ", err)
		return err
	}

	uow.Commit()
	return nil
}

func (s *CustomerService) GetCustomerByID(customerID uint) (*user.Customer, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var customer user.Customer
	if err := s.repository.GetByID(uow, &customer, customerID); err != nil {
		s.repository.Preload("LoginInfo")
		s.log.Error("Error fetching customer: ", err)
		return nil, err
	}

	return &customer, nil
}


