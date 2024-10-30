package service

import (
	"fmt"
	"loanApp/models/loanapplication"
	"loanApp/models/loanscheme"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"strconv"

	"github.com/jinzhu/gorm"
)

type LoanSchemeService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoanSchemeService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoanSchemeService {
	return &LoanSchemeService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (s *LoanSchemeService) CreateLoanScheme(scheme *loanscheme.LoanScheme, userID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()



	var admin user.Admin
	err:=s.repository.GetByID(uow,&admin,userID)
	if err!=nil{
		return err
	}
	scheme.CreatedBy=&admin

	var existingScheme loanscheme.LoanScheme
	err = s.repository.GetAll(uow, &existingScheme, s.repository.Filter("name = ?", scheme.Name))
	if err != nil && !gorm.IsRecordNotFoundError(err) { 
		return fmt.Errorf("error checking for duplicate loan scheme: %v", err)
	}
	if existingScheme.Name != "" {
		return fmt.Errorf("a loan scheme with this name already exists")
	}

	err = s.repository.Add(uow, scheme)
	if err != nil {
		return fmt.Errorf("could not create loan scheme: %v", err)
	}

	uow.Commit()
	return nil
}


func (s *LoanSchemeService) GetAllLoanSchemes(allSchemes *[]*loanscheme.LoanScheme, totalCount *int, parser web.Parser, userID uint ) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 10 // default value
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 // default value
	}

	queryProcessors := []repository.QueryProcessor{
		s.repository.Limit(limit),
		s.repository.Offset(offset),
	}

	if err := s.repository.GetAll(uow, allSchemes, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanSchemeService) UpdateLoanScheme(id string, updatedScheme *loanscheme.LoanScheme, userID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()



	var scheme loanscheme.LoanScheme
	if err := s.repository.GetByID(uow, &scheme, id); err != nil {
		return err
	}

	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(uow, &applications,
		s.repository.Filter("loan_scheme_id = ?", id),
		s.repository.Filter("status IN (?, ?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded", "Approved"))
	if err != nil  {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	if len(applications) > 0 {
		return fmt.Errorf("cannot update loan scheme with ID %s: there are pending applications", id)
	}

	scheme.Name = updatedScheme.Name
	scheme.Category = updatedScheme.Category
	scheme.InterestRate = updatedScheme.InterestRate
	scheme.Tenure = updatedScheme.Tenure

	if err := s.repository.Update(uow, &scheme); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanSchemeService) DeleteLoanScheme(id string, userID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var scheme loanscheme.LoanScheme
	if err := s.repository.GetByID(uow, &scheme, id); err != nil {
		return err
	}

	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(uow, &applications,
		s.repository.Filter("loan_scheme_id = ?", id),
		s.repository.Filter("status IN (?, ?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded", "Approved"))
	if err != nil {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	if len(applications) > 0 {
		return fmt.Errorf("cannot delete loan scheme with ID %s: there are pending applications", id)
	}

	if err := s.repository.DeleteByID(uow, &scheme, id); err != nil {
		return err
	}

	uow.Commit()
	return nil
}
