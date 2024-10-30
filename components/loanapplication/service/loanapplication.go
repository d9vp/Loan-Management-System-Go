package service

import (
	"errors"
	"fmt"
	"io"
	"loanApp/components/loanofficer/service"
	"loanApp/models/document"
	"loanApp/models/installation"
	"loanApp/models/loanapplication"
	"loanApp/models/user"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"loanApp/repository"
	"loanApp/utils/log"

	"gopkg.in/gomail.v2"

	"github.com/jinzhu/gorm"
)

type LoanApplicationService struct {
	DB                 *gorm.DB
	repository         repository.Repository
	log                log.Logger
	LoanOfficerService *service.LoanOfficerService
}

func NewLoanApplicationService(db *gorm.DB, repository repository.Repository, log log.Logger, los *service.LoanOfficerService) *LoanApplicationService {
	return &LoanApplicationService{
		DB:                 db,
		repository:         repository,
		log:                log,
		LoanOfficerService: los,
	}
}

const DocumentUploadDir = "C:\\Users\\Dev.patel\\Downloads\\Loan Management System\\upload"

func (s *LoanApplicationService) GetLoanApplicationsByCustomer(customerID uint, applications *[]loanapplication.LoanApplication) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	queryProcessors := []repository.QueryProcessor{
		s.repository.Filter("customer_id = ?", customerID),
		s.repository.Preload("Documents"),
		s.repository.Preload("Installations"),
	}

	if err := s.repository.GetAll(uow, applications, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) ApplyForLoan(customerID uint, loanSchemeID uint, amount float64, files []*multipart.FileHeader) (uint, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var count int64
	err := s.DB.Model(&loanapplication.LoanApplication{}).
		Where("customer_id = ? AND status IN (?, ?, ?)", customerID, "Pending", "PendingCollateral", "Collateral Uploaded").
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to check existing applications: %w", err)
	}

	if count >= 3 {
		return 0, fmt.Errorf("user can only have up to 3 pending loan applications at a time")
	}

	application := loanapplication.LoanApplication{
		CustomerID:      customerID,
		LoanSchemeID:    loanSchemeID,
		Amount:          amount,
		Status:          "Pending",
		ApplicationDate: time.Now(),
	}

	officer, err := s.LoanOfficerService.GetLeastLoadedOfficer()
	if err != nil {
		return 0, err
	}
	application.LoanOfficerID = officer.ID

	if err := s.repository.Add(uow, &application); err != nil {
		return 0, err
	}

	for _, fileHeader := range files {
		doc, err := s.saveDocument(fileHeader, "personal_documents", application.ID)
		if err != nil {
			return 0, err
		}
		if err := s.repository.Add(uow, doc); err != nil {
			return 0, err
		}
	}

	uow.Commit()
	return application.ID, nil
}

func (s *LoanApplicationService) UploadCollateralDocuments(applicationID string, customerID uint, files []*multipart.FileHeader) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	s.log.Info("upload collateral service called")

	canUpload, err := s.CanUploadCollateral(applicationID, customerID)
	if err != nil || !canUpload {
		return err
	}

	s.log.Info("can upload collateral checked")

	var application loanapplication.LoanApplication
	if err := s.repository.GetByID(uow, &application, applicationID); err != nil {
		return fmt.Errorf("failed to retrieve application: %w", err)
	}

	s.log.Info("retrieved application:", application)

	for _, fileHeader := range files {
		doc, err := s.saveDocument(fileHeader, "collateral_documents", application.ID)
		if err != nil {
			return err
		}

		s.log.Info("saved doc:", doc.DocumentType)

		if err := s.repository.Add(uow, doc); err != nil {
			return err
		}

		s.log.Info("added to db doc:", doc.DocumentType)

	}

	application.Status = "Collateral Uploaded"
	if err := s.repository.Update(uow, &application); err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	s.log.Info("updated lioan applications table:", application)

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) CanUploadCollateral(applicationID string, customerID uint) (bool, error) {
	var application loanapplication.LoanApplication
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	s.log.Info("where is traas")

	applicationIDint, err := strconv.Atoi(applicationID)
	if err != nil {
		return false, err
	}

	s.log.Info("where is traas")

	err = s.repository.GetByID(uow, &application, applicationIDint,
		s.repository.Filter("customer_id = ?", customerID),
		s.repository.Filter("status = ?", "PendingCollateral"),
	)
	if err != nil {
		return false, err
	}
	s.log.Info("where is traas")

	// return application.Status == "Approved", nil
	return true, nil

}

func (s *LoanApplicationService) saveDocument(fileHeader *multipart.FileHeader, docType string, applicationID uint) (*document.Document, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileName := fileHeader.Filename + strconv.Itoa(int(applicationID))
	filePath := filepath.Join(DocumentUploadDir, docType, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	return &document.Document{
		DocumentType:      docType,
		URL:               filePath,
		LoanApplicationID: applicationID,
	}, nil
}

func (s *LoanApplicationService) UpdateLoanApplicationStatus(loanApplicationID uint, loanOfficerID uint, newStatus string) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var application loanapplication.LoanApplication
	if err := s.repository.GetByID(uow, &application, loanApplicationID); err != nil {
		return err
	}

	if application.LoanOfficerID != loanOfficerID {
		return errors.New("loan officer not authorized to update this application")
	}

	if application.Status != newStatus {
		application.Status = newStatus
		*application.DecisionDate = time.Now()
	}

	if err := s.repository.Update(uow, &application); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) PayInstallment(customerID, loanApplicationID uint) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	var installment installation.Installation
	err := s.repository.GetAll(uow, &installment,
		s.repository.Filter("loan_application_id = ? AND status = ?", loanApplicationID, "Pending"),
		s.repository.OrderBy("due_date ASC"),
		s.repository.Limit(1),
	)
	fmt.Println("installments:", installment)

	if err != nil || installment.ID == 0 {
		var loanapp loanapplication.LoanApplication

		err := s.repository.GetByID(uow, &loanapp, loanApplicationID)
		if err != nil {
			return errors.New("loan application not found")
		}
		if loanapp.CustomerID != customerID {
			return errors.New("not this customer's loan application")
		}

		loanapp.Status = "Paid Off"
		if err := s.repository.Update(uow, &loanapp); err != nil {
			return errors.New("failed to update loan application status to paid off")
		}
		uow.Commit()
		return errors.New("no pending installments found for this loan application")

	}

	installment.Status = "Paid"
	now := time.Now()
	installment.PaymentDate = &now

	if err := s.repository.Update(uow, &installment); err != nil {
		return errors.New("failed to update installment payment status")
	}

	uow.Commit()
	return nil
}

func (s *LoanApplicationService) CheckForNPA() error {
	var applications []loanapplication.LoanApplication

	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	if err := s.repository.GetAll(uow, &applications); err != nil {
		return fmt.Errorf("failed to fetch loan applications: %w", err)
	}

	now := time.Now()

	for _, app := range applications {
		var installments []installation.Installation
		err := s.repository.GetAll(uow, &installments,
			s.repository.Filter("loan_application_id = ?", app.ID),
			s.repository.Filter("status = ?", "Pending"),
			s.repository.Filter("due_date < ?", now),
			s.repository.OrderBy("due_date DESC"),
			s.repository.Limit(3),
		)

		if err != nil {
			return fmt.Errorf("failed to fetch installments: %w", err)
		}

		overdueCount := 0
		for _, inst := range installments {
			if inst.DueDate.Before(now) && inst.Status == "Pending" {
				overdueCount++
			}
		}

		if overdueCount == 3 && !app.IsNPA {
			app.IsNPA = true
			if err := s.repository.Update(uow, &app); err != nil {
				return fmt.Errorf("failed to update loan application NPA status: %w", err)
			}
		} else if overdueCount < 3 && app.IsNPA {
			app.IsNPA = false
			if err := s.repository.Update(uow, &app); err != nil {
				return fmt.Errorf("failed to update loan application NPA status: %w", err)
			}
		}
	}

	uow.Commit()
	return nil
}

func ScheduleNPAStatusCheck(service *LoanApplicationService) {
	ticker := time.NewTicker(1 * time.Minute) // Run daily
	service.log.Info("GOROUTINE TO CHECK NPA STATUS")
	go func() {
		for range ticker.C {
			err := service.CheckForNPA()
			if err != nil {
				log.GetLogger().Error("Error running NPA check: %v", err)
			}
		}
	}()
}

func ScheduleReminders(s *LoanApplicationService) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	dueDate := time.Now()

	startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var installments []installation.Installation
	err := s.repository.GetAll(uow, &installments,
		s.repository.Filter("status = ?", "Pending"),
		s.repository.Filter("due_date >= ? AND due_date < ?", startOfDay, endOfDay),
	)
	if err != nil {
		return fmt.Errorf("failed to fetch due installments: %w", err)
	}

	for _, installment := range installments {
		if err := s.sendReminderEmail(installment); err != nil {
			fmt.Printf("failed to send reminder for installment %d: %v\n", installment.ID, err)
		}
	}

	return nil
}

func (s *LoanApplicationService) sendReminderEmail(installment installation.Installation) error {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()
	var customer user.Customer
	fmt.Println("send reminder email")
	var loanapplication loanapplication.LoanApplication
	err := s.repository.GetByID(uow, &loanapplication, installment.LoanApplicationID)
	if err != nil {
		return fmt.Errorf("failed to fetch loan application: %w", err)
	}
	if err := s.repository.GetByID(uow, &customer, loanapplication.CustomerID); err != nil {
		return fmt.Errorf("failed to fetch customer: %w", err)
	}

	subject := "Loan Payment Reminder"
	message := fmt.Sprintf(`
        Dear %s,

        This is a reminder that your installment of $%.2f is due on %s.

        Loan Details:
        Loan Application ID: %d
        Amount Due: $%.2f
        Due Date: %s

        Paisa dedo guis.

        Best regards,
        Loanleloplis
    `, customer.Name, installment.AmountToBePaid, installment.DueDate.Format("2006-01-02"), installment.LoanApplicationID, installment.AmountToBePaid, installment.DueDate.Format("2006-01-02"))

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

func StartReminderScheduler(s *LoanApplicationService) {
	// now := time.Now()
	// nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	// durationUntilMidnight := nextMidnight.Sub(now)

	// Start a goroutine to wait until midnight, then run every 24 hours
	go func() {
		for {
			if err := ScheduleReminders(s); err != nil {
				log.GetLogger().Error("Error scheduling reminders: %v", err)
			}
			time.Sleep(2 * time.Minute) //every 2 min currently
			// time.Sleep(24 * time.Hour) // Sleep for 24 hours
		}
	}()
}
