package modules

import (
	"loanApp/app"
	"loanApp/components/loanapplication/controller"
	"loanApp/components/loanapplication/service"
	los "loanApp/components/loanofficer/service"
)

func registerLoanApplicationRoutes(appObj *app.App) {
	loanofficerService := los.NewLoanOfficerService(appObj.DB, appObj.Repository, appObj.Log)
	loanapplicationService := service.NewLoanApplicationService(appObj.DB, appObj.Repository, appObj.Log, loanofficerService)
	loanofficerController := controller.NewLoanApplicationController(loanapplicationService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loanofficerController})
}
