package modules

import (
	"loanApp/app"

	"loanApp/models/document"
	"loanApp/models/installation"
	"loanApp/models/loanapplication"
	"loanApp/models/loanscheme"
	"loanApp/models/logininfo"
	"loanApp/models/user"
)

func ConfigureAppTables(appObj *app.App) {
	userModuleConfig := user.NewUserModuleConfig(appObj.DB, appObj.Log)
	loginModuleConfig := logininfo.NewLoginInfoModuleConfig(appObj.DB, appObj.Log)
	loanSchemeModuleConfig:=loanscheme.NewLoanSchemeModuleConfig(appObj.DB, appObj.Log)
	loanApplicationModuleConfig:=loanapplication.NewLoanApplicationModuleConfig(appObj.DB,appObj.Log)
	documentModuleConfig:=document.NewDocumentModuleConfig(appObj.DB,appObj.Log)
	installationModuleCofig:=installation.NewInstallationModuleConfig(appObj.DB,appObj.Log)

	appObj.TableMigration([]app.ModuleConfig{userModuleConfig, loginModuleConfig, loanSchemeModuleConfig, loanApplicationModuleConfig, documentModuleConfig,installationModuleCofig})
}
