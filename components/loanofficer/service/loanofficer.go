package service

import (
	"errors"
	"fmt"
	"loanApp/models/installation"
	"loanApp/models/loanapplication"
	"loanApp/models/loanscheme"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/gomail.v2"
)

type LoanOfficerService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewLoanOfficerService(db *gorm.DB, repository repository.Repository, log log.Logger) *LoanOfficerService {
	return &LoanOfficerService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}


func (s *LoanOfficerService) CreateLoanOfficer(officer *user.LoanOfficer, adminID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()



	err := s.repository.Add(uow, &officer.User)
	if err != nil {
		return fmt.Errorf("error creating user data for loan officer: %v", err)
	}

	officer.ID = officer.User.ID

	err = s.repository.Add(uow, officer)
	if err != nil {
		return fmt.Errorf("error creating loan officer: %v", err)
	}

	uow.Commit()

	return nil
}


func (s *LoanOfficerService) GetAllLoanOfficers(allOfficers *[]*user.LoanOfficer, totalCount *int, parser web.Parser, adminID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 12 // default
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 // default
	}

	queryProcessors := []repository.QueryProcessor{
		s.repository.Preload("LoginInfo"),
		s.repository.Preload("AssignedLoans"),
		s.repository.Limit(limit),
		s.repository.Offset(offset),
	}

	if err := s.repository.GetAll(uow, allOfficers, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) UpdateLoanOfficer(id string, updatedOfficer *user.LoanOfficer, adminID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var officer user.LoanOfficer
	if err := s.repository.GetByID(uow, &officer, id); err != nil {
		return err
	}

	officer.Name = updatedOfficer.Name
	officer.Email = updatedOfficer.Email
	officer.Password = updatedOfficer.Password

	if err := s.repository.Update(uow, &officer); err != nil {
		return err
	}
	if err := s.repository.Update(uow, &officer.User); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) DeleteLoanOfficer(id string, adminID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()


	var officer user.LoanOfficer
	if err := s.repository.GetByID(uow, &officer, id); err != nil {
		return err
	}

	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(uow, &applications,
		s.repository.Filter("loan_officer_id = ?", officer.ID),
		s.repository.Filter("status IN (?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded"))
	if err != nil {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	if len(applications) > 0 {
		return fmt.Errorf("cannot delete loan officer with ID %s: there are pending applications", id)
	}

	if err := s.repository.DeleteByID(uow, &officer, id); err != nil {
		return err
	}

	var user user.User
	if err := s.repository.GetByID(uow, &user, id); err != nil {
		return err
	}

	if err := s.repository.DeleteByID(uow, &user, id); err != nil {
		return err
	}

	uow.Commit()
	return nil
}



func (s *LoanOfficerService) GetLeastLoadedOfficer() (*user.LoanOfficer, error) {
	var officer user.LoanOfficer
	err := s.DB.Model(&user.LoanOfficer{}).
		Joins("LEFT JOIN loan_applications ON loan_officers.id = loan_applications.loan_officer_id AND loan_applications.status IN (?, ?, ?)", "Pending", "PendingCollateral", "Collateral Uploaded").
		Where("loan_officers.is_active = ?", true).
		Group("loan_officers.id").
		Order("COUNT(loan_applications.id) ASC").
		First(&officer).Error

	if err != nil {
		return nil, err
	}

	if officer.ID == 0 {
		return nil, fmt.Errorf("no active loan officer found")
	}

	return &officer, nil
}

func (s *LoanOfficerService) GetAssignedLoanApplications(loanOfficerID uint) ([]*loanapplication.LoanApplication, error) {
	var applications []*loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	queryProcessors := []repository.QueryProcessor{
		s.repository.Filter("loan_officer_id = ?", loanOfficerID),
		s.repository.Preload("Installations"),
		s.repository.Preload("Documents"),
	}

	err := s.repository.GetAll(uow, &applications, queryProcessors...)
	if err != nil {
		return nil, err
	}

	return applications, nil
}
func (s *LoanOfficerService) ApproveInitialApplication(applicationID string, loanOfficerID uint, approve bool) error {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	if application.LoanOfficerID!=loanOfficerID{
		return errors.New("this application is assigned to another officer")
	}

	if application.Status == "Approved" || application.Status == "Rejected" || application.Status == "PendingCollateral" {
		return errors.New("invalid operation: application has already been processed")
	}

	if approve {
		nowtime := time.Now()
		application.DecisionDate = &nowtime
		application.Status = "PendingCollateral"

		if err := s.sendApprovalEmail(&application); err != nil {
			return fmt.Errorf("failed to send approval email: %w", err)
		}
	} else {
		application.Status = "Rejected"
	}

	if err := s.repository.Update(uow, &application); err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) sendApprovalEmail(application *loanapplication.LoanApplication) error {

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send approval email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Approval, Pending Collateral"
	message := fmt.Sprintf(`
        Dear %s
		Your loan application (ID: %d) has been approved. Please upload your collateral documents within one week.
		If the documents are not uploaded within the stipulated time, your application will be rejected.
		Thank you for your attention.

        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd")

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *LoanOfficerService) CheckPendingCollateralApplications() error {
	var applications []loanapplication.LoanApplication
	err := s.repository.GetAll(nil, &applications,
		s.repository.Filter("status = ?", "PendingCollateral"))
	if err != nil {
		return fmt.Errorf("failed to fetch pending collateral applications: %w", err)
	}

	now := time.Now()
	for _, application := range applications {
		if application.DecisionDate != nil && application.DecisionDate.Add(7*24*time.Hour).Before(now) {
			application.Status = "Rejected"
			if err := s.repository.Update(nil, &application); err != nil {
				return fmt.Errorf("failed to reject application due to missing collateral: %w", err)
			}
			if err := s.sendRejectionEmail(&application); err != nil {
				return fmt.Errorf("failed to send rejection email: %w", err)
			}
		}
	}

	return nil
}

func (s *LoanOfficerService) sendRejectionEmail(application *loanapplication.LoanApplication) error {

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send rejection email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	// Email content
	subject := "Loan Rejection due to Pending Collateral"
	message := fmt.Sprintf(`
        Dear %s
		Your loan application (ID: %d) has been rejected since you did not upload collateral documents within one week.
		You may apply for another loan if you are still interested.
		Thank you for your attention.

        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd")

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func SchedulePendingCollateralCheck(service *LoanOfficerService) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		for range ticker.C {
			if err := service.CheckPendingCollateralApplications(); err != nil {
				service.log.Error("Error checking pending collateral applications: %v", err)
			}
		}
	}()
}

func (s *LoanOfficerService) ApproveCollateralDocuments(applicationID string, loanOfficerID uint, approve bool) error {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	if application.LoanOfficerID!=loanOfficerID{
		return errors.New("application is assigned to another loan officer")
	}

	if application.Status != "Collateral Uploaded" {
		return errors.New("no collateral documents. application status is not collateral uploaded")
	}

	if approve {
		application.Status = "Approved"
		decisionDate := time.Now()
		application.DecisionDate = &decisionDate

		if err := s.generateInstallments(&application, uow); err != nil {
			return fmt.Errorf("failed to generate installments: %w", err)
		}
	} else {
		application.Status = "Rejected"
	}

	if err := s.repository.Update(uow, &application); err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}
	if err := s.sendFinalApprovalEmail(&application); err != nil {
		return fmt.Errorf("failed to send final approval email: %w", err)
	}

	uow.Commit()
	return nil
}

func (s *LoanOfficerService) sendFinalApprovalEmail(application *loanapplication.LoanApplication) error {

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send final approval email")

	if err := s.repository.GetByID(uow, &customer, application.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	subject := "Loan Approved!"
	message := fmt.Sprintf(`
        Dear %s
		Congratulations! Your loan application (ID: %d) has been approved.
		You can login to check your installment schedule and applications.
		Thank you for your attention.

        Best regards,
        Loanleloplis
    `, customer.Name, application.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "kierarieger2@gmail.com")
	m.SetHeader("To", customer.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	d := gomail.NewDialer("smtp.gmail.com", 587, "kierarieger2@gmail.com", "rttw twcm ponf rbtd")

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *LoanOfficerService) generateInstallments(application *loanapplication.LoanApplication, uow *repository.UOW) error {
	var loanScheme loanscheme.LoanScheme

	if err := s.repository.GetByID(uow, &loanScheme, application.LoanSchemeID); err != nil {
		return fmt.Errorf("failed to fetch loan scheme: %w", err)
	}
	if loanScheme.ID == 0 {
		return fmt.Errorf("loan scheme not found for ID: %d", application.LoanSchemeID)
	}

	monthlyRate := loanScheme.InterestRate / 12 / 100
	emi := (application.Amount * monthlyRate) / (1 - math.Pow(1+monthlyRate, -float64(loanScheme.Tenure)))

	// installments := make([]*installation.Installation, loanScheme.Tenure)
	// for i := 0; i < loanScheme.Tenure; i++ {
	// 	installmentDate := time.Now().AddDate(0, i+1, 0)
	// 	dueDate := time.Date(
	// 		installmentDate.Year(),
	// 		installmentDate.Month(),
	// 		installmentDate.Day(),
	// 		0, 0, 0, 0,
	// 		installmentDate.Location(),
	// 	)

	// 	installments[i] = &installation.Installation{
	// 		LoanApplicationID: application.ID,
	// 		AmountToBePaid:    emi,
	// 		DueDate:           dueDate,
	// 		Status:            "Pending",
	// 	}

	// 	if err := s.repository.Add(uow, installments[i]); err != nil {
	// 		return fmt.Errorf("failed to save installment: %w", err)
	// 	}
	// }

	installments := make([]*installation.Installation, loanScheme.Tenure)
	for i := 0; i < loanScheme.Tenure; i++ {
		dueDate := time.Now().Add(time.Duration(i+1) * time.Minute)

		installments[i] = &installation.Installation{
			LoanApplicationID: application.ID,
			AmountToBePaid:    emi,
			DueDate:           dueDate,
			Status:            "Pending",
		}

		if err := s.repository.Add(uow, installments[i]); err != nil {
			return fmt.Errorf("failed to save installment: %w", err)
		}
	}

	return nil
}
func (s *LoanOfficerService) IsApplicationAssignedToOfficer(applicationID string, loanOfficerID uint) (bool, error) {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return false, fmt.Errorf("failed to retrieve application: %w", err)
	}

	if application.LoanOfficerID == loanOfficerID {
		return true, nil
	}

	return false, nil
}
