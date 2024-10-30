package controller

import (
	"encoding/json"
	"net/http"

	"loanApp/components/customer/service"
	loanschemeservice "loanApp/components/loanscheme/service"

	"loanApp/components/middleware"
	"loanApp/models/loanscheme"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/validation"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type CustomerController struct {
	CustomerService *service.CustomerService
	LoanSchemeService *loanschemeservice.LoanSchemeService
	log             log.Logger
}

func NewCustomerController(customerService *service.CustomerService, log log.Logger, loanschemeService *loanschemeservice.LoanSchemeService ) *CustomerController {
	return &CustomerController{
		CustomerService: customerService,
		LoanSchemeService: loanschemeService,
		log:             log,
	}
}

func (c *CustomerController) RegisterRoutes(router *mux.Router) {
	customerRouter := router.PathPrefix("/customer").Subrouter()
	customerRouter.Use(middleware.TokenAuthMiddleware)
	customerRouter.Use(middleware.CustomerOnly)
	customerRouter.HandleFunc("/update", c.UpdateCustomer).Methods(http.MethodPut)
	customerRouter.HandleFunc("/schemes", c.GetAllLoanSchemes).Methods(http.MethodGet)


}

func (c *CustomerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	c.log.Info("UpdateCustomer called")

	customerID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("Unauthorized access: ", err)
		web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var updatedCustomer user.Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	existingCustomer, err := c.CustomerService.GetCustomerByID(customerID)
	if err != nil {
		c.log.Error("Error fetching customer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch customer")
		return
	}

	if updatedCustomer.Name != "" {
		existingCustomer.Name = updatedCustomer.Name
	}
	if updatedCustomer.Email != "" {
		existingCustomer.Email = updatedCustomer.Email
	}
	if updatedCustomer.Password != "" {
		hashedPassword, err := middleware.HashPassword(updatedCustomer.Password)
		if err != nil {
			c.log.Error("Hashing error: ", err)
			web.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		existingCustomer.Password = hashedPassword
	}

	if err := validation.ValidateEmail(updatedCustomer.Email); err != nil {
		c.log.Error("Email Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.CustomerService.UpdateCustomer(existingCustomer); err != nil {
		c.log.Error("Error updating customer: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update customer")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, existingCustomer)
}

func (c *CustomerController) GetAllLoanSchemes(w http.ResponseWriter, r *http.Request) {

	userID, err := web.GetUserIDFromContext(r)
	if err != nil {
		c.log.Error("No such customer found: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "No customer found")
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
