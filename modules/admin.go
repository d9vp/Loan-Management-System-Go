package modules

import (
	"loanApp/app"
	"loanApp/components/admin/controller"
	"loanApp/components/admin/service"
)

func registerAdminRoutes(appObj *app.App) {
	adminService := service.NewAdminService(appObj.DB, appObj.Repository, appObj.Log)
	adminController := controller.NewAdminController(adminService, appObj.Log)
	appObj.RegisterAllControllerRoutes([]app.Controller{adminController})
}
