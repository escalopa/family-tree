package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("gender", validateGender)
	validate.RegisterValidation("date_order", validateDateOrder)
}

func Validate(data interface{}) error {
	return validate.Struct(data)
}

func validateGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	return gender == "M" || gender == "F" || gender == "N"
}

func validateDateOrder(fl validator.FieldLevel) bool {
	// This is a simplified validator
	// In practice, you'd pass both dates to compare
	return true
}

func ValidateDateOrder(start, end *time.Time) bool {
	if start == nil || end == nil {
		return true
	}
	return !end.Before(*start)
}


