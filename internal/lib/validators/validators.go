package validators

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

type LoginValidator struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	AppID    int    `validate:"required,gt=0"`
}

func (v *LoginValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(v)
}

func ToLoginValidator(email string, password string, appId int) *LoginValidator {
	return &LoginValidator{
		Email:    email,
		Password: password,
		AppID:    appId,
	}
}

type RegisterValidator struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

func (v *RegisterValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(v)
}

func ToRegisterValidator(email string, password string) *RegisterValidator {
	return &RegisterValidator{
		Email:    email,
		Password: password,
	}
}

type IsAdminValidator struct {
	UserID int64 `validate:"required,gt=0"`
}

func (v *IsAdminValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(v)
}

func ToIsAdminValidator(userID int64) *IsAdminValidator {
	return &IsAdminValidator{
		UserID: userID,
	}
}

func GetDetailedError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		firstError := validationErrors[0]

		switch firstError.Tag() {
		case "min":
			return fmt.Sprintf("Field '%s' require minimum %s characters\n", firstError.Tag(), firstError.Param())
		default:
			return fmt.Sprintf("Field '%s' is invalid\n", firstError.Tag())
		}
	}
	return "Internal error"
}
