package controller

import (
	"fmt"
	"net/http"
	"strings"

	logoutservice "loanApp/components/logout/service"
	"loanApp/components/middleware"
	"loanApp/components/user/service"
	"loanApp/models/userclaims"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type LogoutController struct {
	UserService   *service.UserService
	LogoutService *logoutservice.LogoutService
}

func NewLogoutController(userService *service.UserService, logoutService *logoutservice.LogoutService) *LogoutController {
	return &LogoutController{
		UserService:   userService,
		LogoutService: logoutService,
	}
}

func (lc *LogoutController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/logout", lc.LogoutHandler).Methods(http.MethodPost)
}
func (lc *LogoutController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	tok := r.Header.Get("Authorization")
	if tok == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(tok, "Bearer ")

	if middleware.IsTokenBlacklisted(tokenStr) {
		http.Error(w, "Token has already been invalidated", http.StatusUnauthorized)
		return
	}

	claims := &userclaims.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("it'sDevthedev"), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	email := claims.Email
	user, err := lc.UserService.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	middleware.BlacklistToken(tokenStr)

	if err := lc.LogoutService.UpdateLoginInfo(user); err != nil {
		http.Error(w, "Failed to update logout info", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
