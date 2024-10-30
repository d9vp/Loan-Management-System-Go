package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"loanApp/components/logininfo/service"
	"loanApp/models/logininfo"
	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
)

type LoginInfoController struct {
	LoginInfoService *service.LoginInfoService
	log              log.Logger
}

func NewLoginInfoController(service *service.LoginInfoService, log log.Logger) *LoginInfoController {
	return &LoginInfoController{
		LoginInfoService: service,
		log:              log,
	}
}

func (c *LoginInfoController) RegisterRoutes(router *mux.Router) {
	loginRouter := router.PathPrefix("/logininfo").Subrouter()
	loginRouter.HandleFunc("/", c.CreateLoginInfo).Methods(http.MethodPost)
	loginRouter.HandleFunc("/{userID}", c.GetLoginInfo).Methods(http.MethodGet)
	loginRouter.HandleFunc("/logout/{userID}", c.UpdateLogoutTime).Methods(http.MethodPut)
}

func (c *LoginInfoController) CreateLoginInfo(w http.ResponseWriter, r *http.Request) {
	c.log.Info("CreateLoginInfo called")

	var newInfo logininfo.LoginInfo
	if err := json.NewDecoder(r.Body).Decode(&newInfo); err != nil {
		c.log.Error("Invalid input: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if _, err := c.LoginInfoService.CreateLoginInfo(&newInfo); err != nil {
		c.log.Error("Error creating login info: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create login info")
		return
	}

	web.RespondWithJSON(w, http.StatusCreated, newInfo)
}

func (c *LoginInfoController) UpdateLogoutTime(w http.ResponseWriter, r *http.Request) {
	c.log.Info("UpdateLogoutTime called")

	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.log.Error("Invalid user ID: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := c.LoginInfoService.UpdateLogoutTime(uint(userID)); err != nil {
		c.log.Error("Error updating logout time: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not update logout time")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logout time updated successfully"})
}

func (c *LoginInfoController) GetLoginInfo(w http.ResponseWriter, r *http.Request) {
	c.log.Info("GetLoginInfo called")

	userIDStr := mux.Vars(r)["userID"]
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.log.Error("Invalid user ID: ", err)
		web.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	info, err := c.LoginInfoService.GetLoginInfo(uint(userID))
	if err != nil {
		c.log.Error("Error fetching login info: ", err)
		web.RespondWithError(w, http.StatusInternalServerError, "Could not fetch login info")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, info)
}
