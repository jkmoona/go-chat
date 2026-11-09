package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func ErrorJSON(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func MsgJSON(w http.ResponseWriter, status int, key, msg string) {
	WriteJSON(w, status, map[string]string{key: msg})
}

func DecodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return validate.Struct(dst)
}

func SetCookie(w http.ResponseWriter, name, value string, maxAge int, secure, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func ReadCookie(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func BindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		e := ve[0]
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			return field + " is required"
		case "min":
			return field + " must be at least " + e.Param() + " characters"
		case "max":
			return field + " must be at most " + e.Param() + " characters"
		case "alphanum":
			return field + " must contain only letters and numbers"
		case "oneof":
			return field + " must be one of: " + strings.ReplaceAll(e.Param(), " ", ", ")
		case "numeric":
			return field + " must be numeric"
		case "len":
			return field + " must be exactly " + e.Param() + " characters"
		}
	}
	return "invalid request"
}
