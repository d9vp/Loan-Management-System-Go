package modules

import (
	"loanApp/app"
	"loanApp/components/customer/controller"
	"loanApp/components/customer/service"
	loanschemeservice "loanApp/components/loanscheme/service"

)

func registerCustomerRoutes(appObj *app.App) {
	loanschemeService:=loanschemeservice.NewLoanSchemeService(appObj.DB,appObj.Repository,appObj.Log)
	customerService := service.NewCustomerService(appObj.DB, appObj.Repository, appObj.Log)
	customerController := controller.NewCustomerController(customerService, appObj.Log,loanschemeService)
	appObj.RegisterAllControllerRoutes([]app.Controller{customerController})
}
