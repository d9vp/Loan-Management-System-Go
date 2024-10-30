package userclaims

import (
	"loanApp/models/user"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	user.User
	jwt.StandardClaims
}
