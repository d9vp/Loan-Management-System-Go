package main

import (
	"fmt"
	"loanApp/app"
	"loanApp/components/admin/service"
	"loanApp/components/middleware"

	loanappservice "loanApp/components/loanapplication/service"
	loanofficerService "loanApp/components/loanofficer/service"
	"loanApp/models/logininfo"
	"loanApp/models/user"
	"loanApp/modules"
	"loanApp/repository"
	"loanApp/utils/log"
	"sync"
)

func main() {
	logger := log.GetLogger()

	db := app.NewDBConnection(logger)
	if db == nil {
		logger.Error("Db connection failed.")
		return
	}

	var wg sync.WaitGroup

	repository := repository.NewGormRepositoryMySQL()
	adminService := service.NewAdminService(db, repository, logger)
	application := app.NewApp("Loan-Management-System", db, logger, &wg, repository)

	app.ClearDatabase()
	application.Init()

	modules.RegisterAllRoutes(application)
	modules.ConfigureAppTables(application)

	createSuperAdmin(adminService)

	loanOfficerService := loanofficerService.NewLoanOfficerService(db, repository, logger)
	loanApplicationService := loanappservice.NewLoanApplicationService(db, repository, logger, loanOfficerService)

	loanappservice.ScheduleNPAStatusCheck(loanApplicationService)
	loanappservice.StartReminderScheduler(loanApplicationService)
	loanofficerService.SchedulePendingCollateralCheck(loanOfficerService)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := application.StartServer()
		if err != nil {
			fmt.Println(err)
			stopApp(application)
		}
	}()

	wg.Wait()
	stopApp(application)
}

func stopApp(app *app.App) {
	app.StopServer()
	fmt.Println("Application stopped")
}

func createSuperAdmin(adminService *service.AdminService) {
	password, err := middleware.HashPassword("password")
	if err != nil {
		fmt.Println("Hashing error:", err)
		return
	}
	superadmin := &user.Admin{
		User: user.User{
			Name:      "Super Admin",
			Email:     "superadmin@gmail.com",
			Password:  password,
			IsActive:  true,
			Role:      "Admin",
			LoginInfo: []*logininfo.LoginInfo{},
		},
		LoanOfficers: []*user.LoanOfficer{},
	}

	err = adminService.CreateAdmin(superadmin)
	if err != nil {
		fmt.Println("Error creating super admin:", err)
		return
	}

	fmt.Println("Super Admin created successfully")
}
