package repository

import (
	"loanApp/models/loanapplication"
	"loanApp/models/logininfo"
	"loanApp/models/statistics"
	"loanApp/models/user"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

type Repository interface {
	GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error
	GetByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error
	Add(uow *UOW, out interface{}) error
	Update(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error
	DeleteByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error
	Limit(limit interface{}) QueryProcessor
	Offset(offset interface{}) QueryProcessor
	Filter(condition string, args ...interface{}) QueryProcessor
	Count(limit, offset int, totalCount *int) QueryProcessor
	GetUserWithLoginInfo(uow *UOW, userID uint) (*user.User, error)
	Preload(association string) QueryProcessor
	OrderBy(orderCondition string) QueryProcessor

	GetActiveUsersCount(uow *UOW, startDate, endDate time.Time) (int64, error)
	GetTotalLoanApplicationsCount(uow *UOW, startDate, endDate time.Time) (int64, error)
	GetNPACount(uow *UOW, startDate, endDate time.Time) (int64, error)
	GetLoginInfosForCustomers(uow *UOW, startDate, endDate time.Time) ([]logininfo.LoginInfo, error)

	GetLoanSchemeStats(uow *UOW, startDate, endDate time.Time) ([]statistics.LoanSchemeStats, error)
	GetLoanOfficerStats(uow *UOW, startDate, endDate time.Time) ([]statistics.LoanOfficerStats, error)
}

type GormRepositoryMySQL struct{} //implements repository

func NewGormRepositoryMySQL() Repository {
	return &GormRepositoryMySQL{}
}

func executeQueryProcessors(db *gorm.DB, out interface{}, queryProcessors ...QueryProcessor) (*gorm.DB, error) {
	var err error
	for i := 0; i < len(queryProcessors); i++ {
		db, err = queryProcessors[i](db, out)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

func (g *GormRepositoryMySQL) Count(limit, offset int, totalCount *int) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if totalCount != nil {
			err := db.Model(out).Count(totalCount).Error
			if err != nil {
				return db, err
			}
		}
		return db, nil
	}
}

func (g *GormRepositoryMySQL) GetAll(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Find(out).Error
}

func (g *GormRepositoryMySQL) GetByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.First(out, "id = ?", id).Error
}

func (g *GormRepositoryMySQL) Add(uow *UOW, out interface{}) error {
	return uow.DB.Create(out).Error
}

func (g *GormRepositoryMySQL) Update(uow *UOW, out interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Save(out).Error
}

func (g *GormRepositoryMySQL) DeleteByID(uow *UOW, out interface{}, id interface{}, queryProcessors ...QueryProcessor) error {
	db, err := executeQueryProcessors(uow.DB, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(out).Error
}

func (g *GormRepositoryMySQL) Limit(limit interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Limit(limit)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) Offset(offset interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Offset(offset)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args...)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) Preload(association string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Preload(association)
		return db, nil
	}
}

func (g *GormRepositoryMySQL) OrderBy(orderCondition string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Order(orderCondition)
		return db, nil
	}
}

type UOW struct {
	DB       *gorm.DB
	Commited bool
}

func NewUnitOfWork(DB *gorm.DB) *UOW {
	return &UOW{DB: DB.Begin(), Commited: false}
}

func (u *UOW) RollBack() {
	if u.Commited {
		return
	}
	u.DB.Rollback()
}

func (u *UOW) Commit() {
	if u.Commited {
		return
	}
	u.DB.Commit()
	u.Commited = true
}

func (g *GormRepositoryMySQL) GetUserWithLoginInfo(uow *UOW, userID uint) (*user.User, error) {
	var foundUser user.User
	db, err := executeQueryProcessors(uow.DB, &foundUser, g.Preload("LoginInfo"))
	if err != nil {
		return nil, err
	}
	err = db.First(&foundUser, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &foundUser, nil
}

func (g *GormRepositoryMySQL) GetActiveUsersCount(uow *UOW, startDate, endDate time.Time) (int64, error) {
	var count int64
	err := uow.DB.Table("login_infos").
		Joins("JOIN users ON users.id = login_infos.user_id").
		Where("users.is_active = ? AND login_time BETWEEN ? AND ? AND users.role = ?", true, startDate, endDate, "Customer").
		Count(&count).Error
	return count, err
}

func (g *GormRepositoryMySQL) GetTotalLoanApplicationsCount(uow *UOW, startDate, endDate time.Time) (int64, error) {
	var count int64
	err := uow.DB.Model(&loanapplication.LoanApplication{}).
		Where("application_date BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error
	return count, err
}

func (g *GormRepositoryMySQL) GetNPACount(uow *UOW, startDate, endDate time.Time) (int64, error) {
	var count int64
	err := uow.DB.Model(&loanapplication.LoanApplication{}).
		Where("is_npa = ? AND application_date BETWEEN ? AND ?", true, startDate, endDate).
		Count(&count).Error
	return count, err
}

func (g *GormRepositoryMySQL) GetLoginInfosForCustomers(uow *UOW, startDate, endDate time.Time) ([]logininfo.LoginInfo, error) {
	var loginInfos []logininfo.LoginInfo
	err := uow.DB.Table("login_infos").
		Select("login_infos.*").
		Joins("JOIN users ON users.id = login_infos.user_id").
		Where("login_time BETWEEN ? AND ? AND users.role = ?", startDate, endDate, "Customer").
		Find(&loginInfos).Error
	return loginInfos, err
}

func (g *GormRepositoryMySQL) GetLoanSchemeStats(uow *UOW, startDate, endDate time.Time) ([]statistics.LoanSchemeStats, error) {
	var stats []statistics.LoanSchemeStats
	err := uow.DB.Table("loan_applications").
		Select("loan_scheme_id as scheme_id, COUNT(*) as application_count, SUM(CASE WHEN is_npa = 1 THEN 1 ELSE 0 END) as non_performing_asset_in_scheme").
		Where("application_date >= ? AND application_date <= ?", startDate, endDate).
		Group("scheme_id").
		Scan(&stats).Error

	log.Printf("Loan Scheme Stats: %+v, Error: %v", stats, err)
	return stats, err
}

func (g *GormRepositoryMySQL) GetLoanOfficerStats(uow *UOW, startDate, endDate time.Time) ([]statistics.LoanOfficerStats, error) {
	var stats []statistics.LoanOfficerStats
	err := uow.DB.Table("loan_applications").
		Select("loan_officer_id as officer_id, COUNT(*) as application_count, SUM(CASE WHEN is_npa = 1 THEN 1 ELSE 0 END) as non_performing_asset_from_officer").
		Where("application_date BETWEEN ? AND ?", startDate, endDate).
		Group("officer_id").
		Scan(&stats).Error
	return stats, err
}
