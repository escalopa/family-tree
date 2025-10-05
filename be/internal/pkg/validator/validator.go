package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	// TODO: Register custom validators
}

func ValidateStruct(s any) error {
	return validate.Struct(s)
}

func ValidateImageType(filename string) bool {
	// TODO: Implementation
	// Check if file extension is jpeg, jpg, png, or webp
	return false
}

func ValidateImageSize(size int64) bool {
	// Max 3MB
	return size <= 3*1024*1024
}

func ValidateDateOrder(marriageDate, divorceDate *time.Time) bool {
	// TODO: Implementation
	// Ensure marriage_date < divorce_date if both present
	return true
}
