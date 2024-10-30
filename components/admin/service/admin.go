package service

import (
	// "loanApp/app"
	"loanApp/models/statistics"
	"loanApp/models/user"
	"loanApp/repository"
	"loanApp/utils/log"
	"loanApp/utils/web"
	"time"

	"strconv"

	"github.com/jinzhu/gorm"
)

type AdminService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.Logger
}

func NewAdminService(db *gorm.DB, repository repository.Repository, log log.Logger) *AdminService {
	return &AdminService{
		DB:         db,
		repository: repository,
		log:        log,
	}
}

func (u *AdminService) CreateAdmin(newAdmin *user.Admin) error {

	uow := repository.NewUnitOfWork(u.DB)
	defer uow.RollBack()

	

	err := u.repository.Add(uow, &newAdmin.User)
	if err != nil {
		return err
	}

	newAdmin.ID = newAdmin.User.ID
	err = u.repository.Add(uow, newAdmin)
	if err != nil {
		return err
	}

	uow.Commit()
	// app.AllAdmins = append(app.AllAdmins, newAdmin)
	return nil
}



func (u *AdminService) GetAllAdmins(allAdmins *[]*user.Admin, totalCount *int, parser web.Parser) error {
	uow := repository.NewUnitOfWork(u.DB)
	defer uow.RollBack()

	limit, err := strconv.Atoi(parser.Form.Get("limit"))
	if err != nil {
		limit = 12 //default
	}

	offset, err := strconv.Atoi(parser.Form.Get("offset"))
	if err != nil {
		offset = 0 //default
	}

	queryProcessors := []repository.QueryProcessor{
		u.repository.Preload("LoginInfo"),
		u.repository.Preload("LoanOfficers"),

		u.repository.Limit(limit),
		u.repository.Offset(offset),
	}

	if err := u.repository.GetAll(uow, allAdmins, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *AdminService) GetStatistics(startDate, endDate time.Time) (*statistics.Statistics, error) {
	uow := repository.NewUnitOfWork(s.DB)
	defer uow.RollBack()

	activeUsersCount, err := s.repository.GetActiveUsersCount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	loginInfos, err := s.repository.GetLoginInfosForCustomers(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalSessionTime float64
	var sessionCount int
	for _, loginInfo := range loginInfos {
		if loginInfo.LogoutTime != nil {
			totalSessionTime += loginInfo.LogoutTime.Sub(loginInfo.LoginTime).Minutes()
			sessionCount++
		}
	}

	averageSessionTime := 0.0
	if sessionCount > 0 {
		averageSessionTime = totalSessionTime / float64(sessionCount)
	}

	loanApplicationsCount, err := s.repository.GetTotalLoanApplicationsCount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	npaCount, err := s.repository.GetNPACount(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	loanSchemeStats, err := s.repository.GetLoanSchemeStats(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	loanOfficerStats, err := s.repository.GetLoanOfficerStats(uow, startDate, endDate)
	if err != nil {
		return nil, err
	}

	stats := &statistics.Statistics{
		ActiveCustomersCount:       int(activeUsersCount),
		AverageCustomerSessionTime: averageSessionTime,
		TotalLoanApplications:      int(loanApplicationsCount),
		NPACount:                   int(npaCount),
		SchemeStats:                loanSchemeStats,
		OfficerStats:               loanOfficerStats,
	}

	uow.Commit()
	return stats, nil
}
