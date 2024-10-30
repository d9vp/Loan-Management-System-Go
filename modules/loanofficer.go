package modules

import (
	"loanApp/app"
	"loanApp/components/loanofficer/controller"
	"loanApp/components/loanofficer/service"
)

func registerLoanOfficerRoutes(appObj *app.App) {
	loanofficerService := service.NewLoanOfficerService(appObj.DB, appObj.Repository, appObj.Log)
	loanofficerController := controller.NewLoanOfficerController(loanofficerService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loanofficerController})
}
