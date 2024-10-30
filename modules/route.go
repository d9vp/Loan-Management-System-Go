package modules

import "loanApp/app"

func RegisterAllRoutes(app *app.App) {
	registerAdminRoutes(app)
	registerCustomerRoutes(app)
	registerLoanOfficerRoutes(app)
	registerLoanSchemeRoutes(app)
	registerLoanApplicationRoutes(app)
	registerLoginRoutes(app)
	registerLogoutRoutes(app)
	registerRegisterRoutes(app)
}
