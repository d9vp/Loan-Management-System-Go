package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"loanApp/components/loanofficer/service"
	"loanApp/components/middleware"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/validation"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoanOfficerController struct {
	LoanOfficerService *service.LoanOfficerService
	log                log.Logger
}

func NewLoanOfficerController(loanOfficerService *service.LoanOfficerService, log log.Logger) *LoanOfficerController {
	return &LoanOfficerController{
		LoanOfficerService: loanOfficerService,
		log:                log,
	}
}

func (c *LoanOfficerController) RegisterRoutes(router *mux.Router) {
	officerRouter := router.PathPrefix("/loan-officer").Subrouter()
	officerRouter.Use(middleware.TokenAuthMiddleware)
	officerRouter.Use(middleware.AdminOnly)

	officerRouter.HandleFunc("/", c.CreateLoanOfficer).Methods(http.MethodPost)
	officerRouter.HandleFunc("/", c.GetAllLoanOfficers).Methods(http.MethodGet)
	officerRouter.HandleFunc("/{id}", c.UpdateLoanOfficer).Methods(http.MethodPut)
	officerRouter.HandleFunc("/{id}", c.DeleteLoanOfficer).Methods(http.MethodDelete)

	decisionRouter := router.PathPrefix("/loan-officer/decision").Subrouter()
	decisionRouter.Use(middleware.TokenAuthMiddleware)
	decisionRouter.Use(middleware.LoanOfficerOnly)
	decisionRouter.HandleFunc("/applications", c.GetAssignedLoanApplications).Methods(http.MethodGet)

	decisionRouter.HandleFunc("/applications/{id}/initial-approval", c.ApproveInitialApplication).Methods(http.MethodPost)

	decisionRouter.HandleFunc("/applications/{id}/collateral-approval", c.ApproveCollateralDocuments).Methods(http.MethodPost)
}

func (c *LoanOfficerController) CreateLoanOfficer(w http.ResponseWriter, r *http.Request) {
	var newOfficer user.LoanOfficer
	if err := json.NewDecoder(r.Body).Decode(&newOfficer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	if err:=validateLoanOfficer(&newOfficer); err != nil {
		c.log.Error("Input Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if err := validation.ValidateEmail(newOfficer.Email); err != nil {
		c.log.Error("Email Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := middleware.HashPassword(newOfficer.Password)
	if err != nil {
		c.log.Error("Hashing error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	newOfficer.Password = hashedPassword

	newOfficer.CreatedByAdminID = userID //
	newOfficer.Role = "Loan Officer"
	if err := c.LoanOfficerService.CreateLoanOfficer(&newOfficer, newOfficer.CreatedByAdminID); err != nil {
		c.log.Error("Error creating loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newOfficer)

}

func (c *LoanOfficerController) GetAllLoanOfficers(w http.ResponseWriter, r *http.Request) {
	c.log.Info("GetAllLoanOfficers called")

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	parser := web.NewParser(r)
	allLoanOfficers := []*user.LoanOfficer{}
	var totalCount int

	if err := c.LoanOfficerService.GetAllLoanOfficers(&allLoanOfficers, &totalCount, *parser, userID); err != nil {
		c.log.Error("Error fetching officers: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch officers")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, allLoanOfficers)
}

func (c *LoanOfficerController) UpdateLoanOfficer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	officerID := vars["id"]

	var updatedOfficer user.LoanOfficer
	if err := json.NewDecoder(r.Body).Decode(&updatedOfficer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	if err:=validateLoanOfficer(&updatedOfficer); err != nil {
		c.log.Error("Input Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidateEmail(updatedOfficer.Email); err != nil {
		c.log.Error("Email Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := middleware.HashPassword(updatedOfficer.Password)
	if err != nil {
		c.log.Error("Hashing error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	updatedOfficer.Password = hashedPassword

	if err := c.LoanOfficerService.UpdateLoanOfficer(officerID, &updatedOfficer, userID); err != nil {
		c.log.Error("Error updating loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, updatedOfficer)
}

func (c *LoanOfficerController) DeleteLoanOfficer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	officerID := vars["id"]

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	if err := c.LoanOfficerService.DeleteLoanOfficer(officerID, userID); err != nil {
		c.log.Error("Error deleting loan officer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not delete loan officer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, "Loan officer deleted successfully")
}

func (c *LoanOfficerController) GetAssignedLoanApplications(w http.ResponseWriter, r *http.Request) {
	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("User ID not found in context:", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	applications, err := c.LoanOfficerService.GetAssignedLoanApplications(userID)
	if err != nil {
		c.log.Error("Error fetching loan applications:", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan applications")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, applications)
}

func (c *LoanOfficerController) ApproveInitialApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("User ID not found in context:", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var decision struct {
		Approve bool `json:"approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&decision); err != nil {
		c.log.Error("Invalid input:", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	err = c.LoanOfficerService.ApproveInitialApplication(applicationID, userID, decision.Approve)
	if err != nil {
		c.log.Error("Error approving application:", err)
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Application processed successfully"})
}

func (c *LoanOfficerController) ApproveCollateralDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("User ID not found in context:", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var decision struct {
		Approve bool `json:"approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&decision); err != nil {
		c.log.Error("Invalid input:", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	err = c.LoanOfficerService.ApproveCollateralDocuments(applicationID, userID, decision.Approve)
	if err != nil {
		c.log.Error("Error approving collateral:", err)
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Collateral processed successfully"})
}

func validateLoanOfficer(loanOfficer *user.LoanOfficer) error {
	if loanOfficer.Name == "" {
		return errors.New("name cannot be empty")
	}
	if loanOfficer.Email == "" {
		return errors.New("email cannot be empty")
	}
	if loanOfficer.Password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}
