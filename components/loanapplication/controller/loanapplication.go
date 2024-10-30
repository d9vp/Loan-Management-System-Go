package controller

import (
	"errors"
	"net/http"
	"strconv"

	"loanApp/components/loanapplication/service"
	"loanApp/components/middleware"
	"loanApp/models/loanapplication"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoanApplicationController struct {
	LoanAppService *service.LoanApplicationService
	log            log.Logger
}

func NewLoanApplicationController(loanAppService *service.LoanApplicationService, log log.Logger) *LoanApplicationController {
	return &LoanApplicationController{
		LoanAppService: loanAppService,
		log:            log,
	}
}
func (c *LoanApplicationController) RegisterRoutes(router *mux.Router) {
	loanAppRouter := router.PathPrefix("/loan-applications").Subrouter()
	loanAppRouter.Use(middleware.TokenAuthMiddleware)
	loanAppRouter.Use(middleware.CustomerOnly)

	loanAppRouter.HandleFunc("", c.ApplyForLoanWithPersonalDocuments).Methods(http.MethodPost)
	loanAppRouter.HandleFunc("", c.GetCustomerLoanApplications).Methods(http.MethodGet)
	loanAppRouter.HandleFunc("/{id}/upload-collateral-docs", c.UploadCollateralDocuments).Methods(http.MethodPost)
	loanAppRouter.HandleFunc("/{id}/pay", c.PayInstallment).Methods(http.MethodPut)
}

func (c *LoanApplicationController) ApplyForLoanWithPersonalDocuments(w http.ResponseWriter, r *http.Request) {
	c.log.Info("ApplyForLoanWithPersonalDocuments called")

	if err := r.ParseMultipartForm(10 << 20); err != nil { // Limit to 10MB
		web.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	loanSchemeIDStr := r.FormValue("loan_scheme_id")
	amountStr := r.FormValue("amount")
	loanSchemeID, amount, err := parseLoanSchemeAndAmount(loanSchemeIDStr, amountStr)
	if err != nil {
		c.log.Error("Invalid loan scheme ID or amount: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	files := r.MultipartForm.File["personal_documents"]
	if len(files) == 0 {
		web.RespondWithError(w, http.StatusBadRequest, "No personal documents provided")
		return
	}

	applicationID, err := c.LoanAppService.ApplyForLoan(customerID, loanSchemeID, amount, files)
	if err != nil {
		c.log.Error("Error creating loan application: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan application")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":       "Loan application created successfully",
		"applicationID": applicationID,
	})
}
func parseLoanSchemeAndAmount(loanSchemeIDStr, amountStr string) (uint, float64, error) {
	loanSchemeID, err := strconv.Atoi(loanSchemeIDStr)
	if err != nil {
		return 0, 0, errors.New("invalid loan scheme ID")
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return 0, 0, errors.New("invalid loan amount")
	}

	return uint(loanSchemeID), amount, nil
}

func (c *LoanApplicationController) UploadCollateralDocuments(w http.ResponseWriter, r *http.Request) {
	c.log.Info("UploadCollateralDocuments called")

	applicationID := mux.Vars(r)["id"]
	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	files := r.MultipartForm.File["collateral_documents"]
	if len(files) == 0 {
		web.RespondWithError(w, http.StatusBadRequest, "No collateral documents provided")
		return
	}

	c.log.Info("collateral docs:",files)

	err = c.LoanAppService.UploadCollateralDocuments(applicationID, customerID, files)
	if err != nil {
		c.log.Error("Error uploading collateral documents: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not upload collateral documents")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, "Collateral documents uploaded successfully")
}

func (c *LoanApplicationController) GetCustomerLoanApplications(w http.ResponseWriter, r *http.Request) {
	c.log.Info("GetCustomerLoanApplications called")
	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var applications []loanapplication.LoanApplication
	if err := c.LoanAppService.GetLoanApplicationsByCustomer(customerID, &applications); err != nil {
		c.log.Error("Error fetching loan applications: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan applications")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, applications)
}

func (c *LoanApplicationController) PayInstallment(w http.ResponseWriter, r *http.Request) {
	c.log.Info("PayInstallment called")

	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	vars := mux.Vars(r)
	loanAppID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		c.log.Error("Invalid loan application ID: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid loan application ID")
		return
	}

	err = c.LoanAppService.PayInstallment(uint(customerID), uint(loanAppID))
	if err != nil {
		c.log.Error("Error processing installment payment: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Nearest due installment paid successfully"})
}
