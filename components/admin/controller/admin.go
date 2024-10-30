package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"loanApp/components/admin/service"
	"loanApp/components/middleware"
	"loanApp/models/user"
	"loanApp/utils/log"
	"loanApp/utils/validation"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type AdminController struct {
	AdminService *service.AdminService
	log          log.Logger
}

func NewAdminController(AdminService *service.AdminService, log log.Logger) *AdminController {
	return &AdminController{
		AdminService: AdminService,
		log:          log,
	}
}

func (a *AdminController) RegisterRoutes(router *mux.Router) {
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.TokenAuthMiddleware)
	adminRouter.Use(middleware.AdminOnly)
	adminRouter.HandleFunc("/", a.CreateAdmin).Methods(http.MethodPost)
	adminRouter.HandleFunc("/", a.GetAllAdmins).Methods(http.MethodGet)
	adminRouter.HandleFunc("/stats", a.GetStatistics).Methods(http.MethodGet)
}

func (a *AdminController) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	a.log.Info("CreateAdmin called")


	var newAdmin user.Admin
	if err := json.NewDecoder(r.Body).Decode(&newAdmin); err != nil {
		a.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := validateAdmin(newAdmin); err != nil {
		a.log.Error("Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ValidateEmail(newAdmin.Email); err != nil {
		a.log.Error("Email Validation error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	hashedPassword, err := middleware.HashPassword(newAdmin.User.Password)
	fmt.Println(hashedPassword)
	if err != nil {
		a.log.Error("Hashing error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	newAdmin.User.Password = hashedPassword
	newAdmin.Role = "Admin"

	if err := a.AdminService.CreateAdmin(&newAdmin); err != nil {
		a.log.Error("Error creating admin: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create admin")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newAdmin)
}

func (a *AdminController) GetAllAdmins(w http.ResponseWriter, r *http.Request) {
	a.log.Info("GetAllAdmins called")

	parser := web.NewParser(r)
	allAdmins := []*user.Admin{}
	var totalCount int

	if err := a.AdminService.GetAllAdmins(&allAdmins, &totalCount, *parser); err != nil {
		a.log.Error("Error fetching admins: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch admins")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, allAdmins)
}

func validateAdmin(admin user.Admin) error {
	if admin.Name == "" {
		return errors.New("name cannot be empty")
	}
	if admin.Email == "" {
		return errors.New("email cannot be empty")
	}
	if admin.Password == "" {
		return errors.New("password cannot be empty")
	}

	return nil
}

func (a *AdminController) GetStatistics(w http.ResponseWriter, r *http.Request) {
	a.log.Info("GetStatistics called")

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "Invalid start date format, use YYYY-MM-DD")
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "Invalid end date format, use YYYY-MM-DD")
		return
	}

	stats, err := a.AdminService.GetStatistics(startDate, endDate)
	if err != nil {
		a.log.Error("Error fetching statistics: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch statistics")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, stats)
}
