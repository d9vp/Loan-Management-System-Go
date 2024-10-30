package modules

import (
	"loanApp/app"
	"loanApp/components/loanscheme/controller"
	"loanApp/components/loanscheme/service"
)

func registerLoanSchemeRoutes(appObj *app.App) {
	loanSchemeService := service.NewLoanSchemeService(appObj.DB, appObj.Repository, appObj.Log)
	loanSchemeController := controller.NewLoanSchemeController(loanSchemeService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loanSchemeController})
}
