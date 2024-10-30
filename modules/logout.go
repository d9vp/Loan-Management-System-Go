package modules

import (
	"loanApp/app"
	"loanApp/components/logout/controller"
	logoutservice "loanApp/components/logout/service"
	"loanApp/components/user/service"
)

func registerLogoutRoutes(appObj *app.App) {

	userService := service.NewUserService(appObj.DB, appObj.Repository, appObj.Log)
	logoutService := logoutservice.NewLogoutService(appObj.DB, appObj.Repository, appObj.Log)

	logoutController := controller.NewLogoutController(userService, logoutService)
	appObj.RegisterAllControllerRoutes([]app.Controller{logoutController})
}
