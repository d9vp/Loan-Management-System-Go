package modules

import (
	"loanApp/app"
	"loanApp/components/login/controller"
	loginservice "loanApp/components/login/service"
	"loanApp/components/user/service"
)

func registerLoginRoutes(appObj *app.App) {

	userService := service.NewUserService(appObj.DB, appObj.Repository, appObj.Log)
	loginService := loginservice.NewLoginService(appObj.DB, appObj.Repository, appObj.Log)

	loginController := controller.NewLoginController(userService, loginService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{loginController})
}
