package modules

import (
	"loanApp/app"
	"loanApp/components/customer/service"
	"loanApp/components/register/controller"
)

func registerRegisterRoutes(appObj *app.App) {

	customerService := service.NewCustomerService(appObj.DB, appObj.Repository, appObj.Log)

	loginController := controller.NewRegisterController(customerService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loginController})
}
