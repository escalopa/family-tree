package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type ErrorCode string

const (
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidToken      ErrorCode = "INVALID_TOKEN"
	ErrCodeSessionExpired    ErrorCode = "SESSION_EXPIRED"
	ErrCodeInvalidOAuthState ErrorCode = "INVALID_OAUTH_STATE"

	ErrCodeForbidden               ErrorCode = "FORBIDDEN"
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"

	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeVersionConflict ErrorCode = "VERSION_CONFLICT"
	ErrCodeInvalidDate     ErrorCode = "INVALID_DATE"

	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

func (e ErrorCode) String() string {
	return string(e)
}

type DomainError struct {
	Code           ErrorCode
	TranslationKey string
	Params         map[string]string
	Err            error
}

func (e *DomainError) Message() string {
	if len(e.Params) > 0 {
		return fmt.Sprintf("%s: %v", e.TranslationKey, e.Params)
	}
	return e.TranslationKey
}

func (e *DomainError) Error() string {
	msg := e.Message()
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, msg)
}

func (e *DomainError) HTTPStatusCode() int {
	switch e.Code {
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeSessionExpired:
		return http.StatusUnauthorized
	case ErrCodeInvalidOAuthState:
		return http.StatusBadRequest
	case ErrCodeForbidden, ErrCodeInsufficientPermissions:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists, ErrCodeConflict, ErrCodeVersionConflict:
		return http.StatusConflict
	case ErrCodeInvalidInput, ErrCodeInvalidDate:
		return http.StatusBadRequest
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeDatabaseError, ErrCodeInternal:
		return http.StatusInternalServerError
	case ErrCodeExternalService:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func NewUnauthorizedError(translationKey string, err error) *DomainError {
	if translationKey == "" {
		translationKey = "error.unauthorized"
	}
	return &DomainError{
		Code:           ErrCodeUnauthorized,
		TranslationKey: translationKey,
		Err:            err,
	}
}

func NewForbiddenError(translationKey string) *DomainError {
	if translationKey == "" {
		translationKey = "error.forbidden"
	}
	return &DomainError{
		Code:           ErrCodeForbidden,
		TranslationKey: translationKey,
	}
}

func NewNotFoundError(resource string) *DomainError {
	translationKey := fmt.Sprintf("error.%s.not_found", strings.ToLower(resource))
	return &DomainError{
		Code:           ErrCodeNotFound,
		TranslationKey: translationKey,
		Params:         map[string]string{"resource": resource},
	}
}

func NewAlreadyExistsError(resource string) *DomainError {
	translationKey := fmt.Sprintf("error.%s.already_exists", strings.ToLower(resource))
	return &DomainError{
		Code:           ErrCodeAlreadyExists,
		TranslationKey: translationKey,
		Params:         map[string]string{"resource": resource},
	}
}

func NewConflictError(translationKey string, params map[string]string) *DomainError {
	if translationKey == "" {
		translationKey = "error.conflict"
	}
	return &DomainError{
		Code:           ErrCodeConflict,
		TranslationKey: translationKey,
		Params:         params,
	}
}

func NewValidationError(translationKey string) *DomainError {
	if translationKey == "" {
		translationKey = "error.invalid_input"
	}
	return &DomainError{
		Code:           ErrCodeInvalidInput,
		TranslationKey: translationKey,
	}
}

func (e *DomainError) WithParams(params map[string]string) *DomainError {
	e.Params = params
	return e
}

func NewVersionConflictError() *DomainError {
	return &DomainError{
		Code:           ErrCodeVersionConflict,
		TranslationKey: "error.version_conflict",
	}
}

func NewInternalError(err error) *DomainError {
	return &DomainError{
		Code:           ErrCodeInternal,
		TranslationKey: "error.internal",
		Err:            err,
	}
}

func NewDatabaseError(err error) *DomainError {
	return &DomainError{
		Code:           ErrCodeDatabaseError,
		TranslationKey: "error.database",
		Err:            err,
	}
}

func NewExternalServiceError(err error) *DomainError {
	return &DomainError{
		Code:           ErrCodeExternalService,
		TranslationKey: "error.external_service",
		Err:            err,
	}
}

func NewInvalidOAuthStateError() *DomainError {
	return &DomainError{
		Code:           ErrCodeInvalidOAuthState,
		TranslationKey: "error.invalid_oauth_state",
	}
}

func NewRateLimitError() *DomainError {
	return &DomainError{
		Code:           ErrCodeTooManyRequests,
		TranslationKey: "error.rate_limit_exceeded",
	}
}

func IsDomainError(err error, code ErrorCode) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == code
	}
	return false
}
