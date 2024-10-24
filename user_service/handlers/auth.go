package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	commonhttp "smapp/common/http"
	"smapp/user_service/repository"
	"smapp/user_service/service"
	"strings"

	"golang.org/x/crypto/bcrypt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SignupRequestBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Handle   string `json:"handle"`
	ImageURL string `json:"imageURL"`
	Password string `json:"password"`
}

// TODO: use constants for field length limits
func (user *SignupRequestBody) Validate() error {
	return validation.ValidateStruct(
		user,
		validation.Field(&user.Name, validation.Required, validation.Length(1, 50)),
		validation.Field(&user.Email, validation.Required, is.Email),
		validation.Field(
			&user.Handle,
			validation.Required,
			validation.Length(1, 20),
			validation.Match(regexp.MustCompile("^[^@]*$")).Error("handle cannot contain '@'"),
		),
		validation.Field(&user.Password, validation.Required, validation.Length(6, 128)),
	)
}

func Signup(auth *service.Auth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user SignupRequestBody
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			commonhttp.JSONError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = user.Validate()
		if err != nil {
			if e, ok := err.(validation.InternalError); ok {
				log.Println(e.InternalError())
				commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
				return
			}
			errors := (err.(validation.Errors).Filter()).(validation.Errors)
			commonhttp.JSONValidationError(w, errors, http.StatusBadRequest)
			return
		}

		token, err := auth.Signup(r.Context(), user.Name, user.Email, user.Handle, user.Password)
		if errors.Is(err, repository.ErrEmailExists) {
			commonhttp.JSONError(w, "Email already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrHandleExists) {
			commonhttp.JSONError(w, "Handle already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status": "success",
			"data":   map[string]string{"token": token},
		}
		commonhttp.JSONResponse(w, response, http.StatusCreated)
	})
}

type LoginRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// TODO: use constants for field length limits
func (user *LoginRequestBody) Validate() error {
	return validation.ValidateStruct(
		user,
		validation.Field(
			&user.Identifier,
			validation.Required,
			validation.When(
				strings.ContainsRune(user.Identifier, '@'),
				validation.Length(6, 254),
			).Else(
				validation.Length(1, 20),
			),
		),
		validation.Field(&user.Password, validation.Required, validation.Length(6, 128)),
	)
}

func Login(auth *service.Auth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user LoginRequestBody
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			commonhttp.JSONError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		wrongCredentialsMessage := "Incorrect email/handle or password"
		err = user.Validate()
		if err != nil {
			if e, ok := err.(validation.InternalError); ok {
				log.Println(e.InternalError())
				commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
				return
			}
			errors := err.(validation.Errors)
			// Keep only "cannot be blank" errors
			for field, e := range errors {
				code := e.(validation.Error).Code()
				if code != validation.ErrRequired.Code() && code != validation.ErrNilOrNotEmpty.Code() {
					delete(errors, field)
				}
			}
			if len(errors) > 0 {
				commonhttp.JSONValidationError(w, errors, http.StatusBadRequest)
				return
			}
			commonhttp.JSONError(w, wrongCredentialsMessage, http.StatusUnauthorized)
			return
		}

		token, err := auth.Login(r.Context(), user.Identifier, []byte(user.Password))
		if errors.Is(err, repository.ErrNoSuchUser) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			commonhttp.JSONError(w, wrongCredentialsMessage, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status": "success",
			"data":   map[string]string{"token": token},
		}
		commonhttp.JSONResponse(w, response, http.StatusOK)
	})
}