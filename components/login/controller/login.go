package controller

import (
	"encoding/json"
	"fmt"
	"loanApp/models/user"
	"net/http"
	"time"

	loginservice "loanApp/components/login/service"
	"loanApp/components/middleware"
	"loanApp/components/user/service"
	"loanApp/models/logininfo"

	"loanApp/utils/log"
	"loanApp/utils/web"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type LoginController struct {
	log          log.Logger
	UserService  *service.UserService
	LoginService *loginservice.LoginService
}

func NewLoginController(userService *service.UserService, loginService *loginservice.LoginService, log log.Logger) *LoginController {
	return &LoginController{
		log:          log,
		UserService:  userService,
		LoginService: loginService,
	}
}

func (lc *LoginController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", lc.Login).Methods(http.MethodPost)
}

func (lc *LoginController) Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	hashedPassword, err := middleware.HashPassword(credentials.Password)
	if err != nil {
		lc.log.Error("Hashing error: ", err)
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println(hashedPassword)
	credentials.Password = hashedPassword

	user, err := lc.UserService.AuthenticateUser(credentials.Email, credentials.Password)
	if err != nil {
		web.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	activeLogin, err := lc.LoginService.GetActiveLoginInfo(int(user.ID))
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Error checking active login")
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour)

	if activeLogin != nil {
		token, err := lc.createToken(user, expirationTime)
		if err != nil {
			web.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
			return
		}

		web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message":   "User already logged in",
			"token":     token,
			"expiresAt": expirationTime,
			"role":      user.Role,
			"user_id":   user.ID,
		})
		return
	}

	token, err := lc.createToken(user, expirationTime)
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
		return
	}

	info := &logininfo.LoginInfo{UserID: user.ID, LoginTime: time.Now()}
	err = lc.LoginService.CreateLoginInfo(user, info)
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, "Error creating login info")
		return
	}

	web.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":     token,
		"expiresAt": expirationTime,
		"role":      user.Role,
		"user_id":   user.ID,
	})
}

func (lc *LoginController) createToken(u *user.User, expirationTime time.Time) (string, error) {
	claims := &user.User{
		Model:    gorm.Model{ID: u.ID},
		Email:    u.Email,
		Role:     u.Role,
		IsActive: u.IsActive,
	}

	token, err := middleware.SigningToken(claims, expirationTime)
	if err != nil {
		return "", err
	}

	return token, nil
}
