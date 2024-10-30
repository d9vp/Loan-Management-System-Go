package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"loanApp/models/user"
	"loanApp/models/userclaims"
	"loanApp/utils/web"

	"github.com/dgrijalva/jwt-go"
)

var SecretKey = []byte("it'sDevthedev")

func HashPassword(password string) (string, error) {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func SigningToken(u *user.User, expirationTime time.Time) (string, error) {
	claims := &userclaims.UserClaims{
		User: *u,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if IsTokenBlacklisted(tokenStr) {
			http.Error(w, "Token has been invalidated, please log in again", http.StatusUnauthorized)
			return
		}

		claims, err := VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func VerifyJWT(tokenStr string) (*userclaims.UserClaims, error) {
	claims := &userclaims.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Admin" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CustomerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Customer" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoanOfficerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
		if !ok || claims == nil {
			web.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: claims not found")
			return
		}

		if claims.User.Role != "Loan Officer" {
			web.RespondWithError(w, http.StatusForbidden, "Unauthorized access")
			return
		}

		next.ServeHTTP(w, r)
	})
}

var (
	blacklistedTokens = make(map[string]struct{})
	mu                sync.Mutex
)

func BlacklistToken(token string) {
	mu.Lock()
	defer mu.Unlock()
	blacklistedTokens[token] = struct{}{}
}

func IsTokenBlacklisted(token string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, exists := blacklistedTokens[token]
	return exists
}
