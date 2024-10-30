package web

import (
	"encoding/json"
	"errors"
	"io"
	"loanApp/models/userclaims"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func UnMarshalJSON(request *http.Request, out interface{}) error {
	//validation
	return json.NewDecoder(request.Body).Decode(out)

}

type Parser struct {
	Body   io.ReadCloser
	Form   url.Values
	Params map[string]string
}

func NewParser(request *http.Request) *Parser {
	return &Parser{
		Body:   request.Body,
		Form:   request.Form,
		Params: mux.Vars(request),
	}

}

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"error": message})
}

func GetUserIDFromContext(r *http.Request) (uint, error) {
	claims, ok := r.Context().Value("claims").(*userclaims.UserClaims)
	if !ok || claims == nil {
		return 0, errors.New("user claims not found in context")
	}

	return claims.User.ID, nil
}
