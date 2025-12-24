package i18n

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateValidationError converts validator.FieldError to localized message
func TranslateValidationError(err validator.FieldError, lang string) string {
	lang = NormalizeLanguage(lang)

	fieldKey := fmt.Sprintf("field.%s", toSnakeCase(err.Field()))
	fieldName := Translate(fieldKey, lang, nil)
	if fieldName == fieldKey {
		// Fallback to field name itself if no translation
		fieldName = err.Field()
	}

	validationKey := fmt.Sprintf("validation.%s", err.Tag())
	params := map[string]string{
		"field": fieldName,
	}

	switch err.Tag() {
	case "min":
		params["min"] = err.Param()
	case "max":
		params["max"] = err.Param()
	case "len":
		params["len"] = err.Param()
	case "gt", "gte", "lt", "lte":
		params["value"] = err.Param()
	case "oneof":
		params["values"] = strings.ReplaceAll(err.Param(), " ", ", ")
	}

	message := Translate(validationKey, lang, params)

	// If translation not found, return generic message
	if message == validationKey {
		return fmt.Sprintf("%s validation failed: %s", fieldName, err.Tag())
	}

	return message
}

func TranslateValidationErrors(err error, lang string) string {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	var messages []string
	for _, fieldErr := range validationErrs {
		messages = append(messages, TranslateValidationError(fieldErr, lang))
	}

	return strings.Join(messages, "; ")
}

// toSnakeCase converts "DateOfBirth" -> "date_of_birth"
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
