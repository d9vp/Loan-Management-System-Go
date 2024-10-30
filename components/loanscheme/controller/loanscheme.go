package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	// "loanApp/app"
	"loanApp/components/loanscheme/service"
	"loanApp/components/middleware"
	"loanApp/models/loanscheme"
	// "loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoanSchemeController struct {
	LoanSchemeService *service.LoanSchemeService
	log               log.Logger
}

func NewLoanSchemeController(loanSchemeService *service.LoanSchemeService, log log.Logger) *LoanSchemeController {
	return &LoanSchemeController{
		LoanSchemeService: loanSchemeService,
		log:               log,
	}
}

func (c *LoanSchemeController) RegisterRoutes(router *mux.Router) {
	schemeRouter := router.PathPrefix("/loan-scheme").Subrouter()
	schemeRouter.Use(middleware.TokenAuthMiddleware)
	schemeRouter.Use(middleware.AdminOnly)
	schemeRouter.HandleFunc("/", c.CreateLoanScheme).Methods(http.MethodPost)
	schemeRouter.HandleFunc("/", c.GetAllLoanSchemes).Methods(http.MethodGet)
	schemeRouter.HandleFunc("/{id}", c.UpdateLoanScheme).Methods(http.MethodPut)
	schemeRouter.HandleFunc("/{id}", c.DeleteLoanScheme).Methods(http.MethodDelete)
}

func (c *LoanSchemeController) CreateLoanScheme(w http.ResponseWriter, r *http.Request) {
	var newScheme loanscheme.LoanScheme
	if err := json.NewDecoder(r.Body).Decode(&newScheme); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := validateLoanScheme(&newScheme); err != nil {
		c.log.Error("Invalid input values: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input values")
		return
	}

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	// var admin *user.Admin
	// for _, a := range app.AllAdmins {
	// 	if a.ID == userID {
	// 		admin = a
	// 		break
	// 	}
	// }
	// if admin == nil {
	// 	c.log.Error("Admin user not found")
	// 	web.RespondWithError(w, http.StatusForbidden, "Admin privileges required")
	// 	return
	// }


	// newScheme.CreatedBy = admin
	newScheme.AdminID = userID

	if err := c.LoanSchemeService.CreateLoanScheme(&newScheme,newScheme.AdminID); err != nil {
		c.log.Error("Error creating loan scheme: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create loan scheme")
		return
	}

	// app.AllLoanSchemes = append(app.AllLoanSchemes, &newScheme)

	web.RespondWithJSON(w, http.StatusCreated, newScheme)
}

func (c *LoanSchemeController) GetAllLoanSchemes(w http.ResponseWriter, r *http.Request) {

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}
	var allSchemes []*loanscheme.LoanScheme
	totalCount := 0

	parser := web.NewParser(r)



	if err := c.LoanSchemeService.GetAllLoanSchemes(&allSchemes, &totalCount, *parser, userID); err != nil {
		c.log.Error("Error fetching loan schemes: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch loan schemes")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, allSchemes)
}

func (c *LoanSchemeController) UpdateLoanScheme(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemeID := vars["id"]

	var updatedScheme loanscheme.LoanScheme
	if err := json.NewDecoder(r.Body).Decode(&updatedScheme); err != nil {
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

	// var admin *user.Admin
	// for _, a := range app.AllAdmins {
	// 	if a.ID == userID {
	// 		admin = a
	// 		break
	// 	}
	// }
	// if admin == nil {
	// 	c.log.Error("Admin user not found")
	// 	web.RespondWithError(w, http.StatusForbidden, "Admin privileges required")
	// 	return
	// }

	if err := validateLoanScheme(&updatedScheme); err != nil {
		c.log.Error("Invalid input values: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input values")
		return
	}

	// updatedScheme.UpdatedBy = append(updatedScheme.UpdatedBy, admin)

	if err := c.LoanSchemeService.UpdateLoanScheme(schemeID, &updatedScheme,userID); err != nil {
		c.log.Error("Error updating loan scheme: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update loan scheme")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, updatedScheme)
}

func (c *LoanSchemeController) DeleteLoanScheme(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemeID := vars["id"]

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such admin found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No admin found")
		return
	}

	if err := c.LoanSchemeService.DeleteLoanScheme(schemeID,userID); err != nil {
		c.log.Error("Error deleting loan scheme: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not delete loan scheme")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, "Loan scheme deleted successfully")
}

func validateLoanScheme(scheme *loanscheme.LoanScheme) error {
	if scheme.Name == "" {
		return errors.New("scheme name cannot be empty")
	}
	if scheme.Category == "" || (scheme.Category != "retail" && scheme.Category != "corporate") {
		return errors.New("category cannot be empty. Should be corporate or retail only")
	}
	if scheme.InterestRate <= 0 {
		return errors.New("interest rate must be greater than zero")
	}
	if scheme.Tenure <= 3 {
		return errors.New("tenure must be greater than three months")
	}
	return nil
}
