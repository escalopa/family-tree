package domain

import (
	"errors"
	"fmt"
)

// Domain errors for consistent error handling
type ErrorCode string

const (
	// Authentication errors
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidToken      ErrorCode = "INVALID_TOKEN"
	ErrCodeSessionExpired    ErrorCode = "SESSION_EXPIRED"
	ErrCodeInvalidOAuthState ErrorCode = "INVALID_OAUTH_STATE"

	// Authorization errors
	ErrCodeForbidden               ErrorCode = "FORBIDDEN"
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"

	// Resource errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// Validation errors
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeVersionConflict ErrorCode = "VERSION_CONFLICT"
	ErrCodeInvalidDate     ErrorCode = "INVALID_DATE"

	// Internal errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// Sentinel errors for errors.Is comparison
var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrConflict          = errors.New("conflict")
	ErrInvalidInput      = errors.New("invalid input")
	ErrVersionConflict   = errors.New("version conflict")
	ErrInvalidOAuthState = errors.New("invalid oauth state")
	ErrInternal          = errors.New("internal error")
	ErrDatabase          = errors.New("database error")
	ErrExternalService   = errors.New("external service error")
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is comparison
func (e *DomainError) Is(target error) bool {
	switch e.Code {
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeSessionExpired:
		return errors.Is(target, ErrUnauthorized)
	case ErrCodeForbidden, ErrCodeInsufficientPermissions:
		return errors.Is(target, ErrForbidden)
	case ErrCodeNotFound:
		return errors.Is(target, ErrNotFound)
	case ErrCodeAlreadyExists:
		return errors.Is(target, ErrAlreadyExists)
	case ErrCodeConflict:
		return errors.Is(target, ErrConflict)
	case ErrCodeInvalidInput:
		return errors.Is(target, ErrInvalidInput)
	case ErrCodeVersionConflict:
		return errors.Is(target, ErrVersionConflict)
	case ErrCodeInvalidOAuthState:
		return errors.Is(target, ErrInvalidOAuthState)
	case ErrCodeDatabaseError:
		return errors.Is(target, ErrDatabase)
	case ErrCodeExternalService:
		return errors.Is(target, ErrExternalService)
	case ErrCodeInternal:
		return errors.Is(target, ErrInternal)
	}
	return false
}

// Error constructors
func NewUnauthorizedError(message string, err error) *DomainError {
	return &DomainError{Code: ErrCodeUnauthorized, Message: message, Err: err}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{Code: ErrCodeForbidden, Message: message}
}

func NewNotFoundError(resource string) *DomainError {
	return &DomainError{Code: ErrCodeNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

func NewAlreadyExistsError(resource string) *DomainError {
	return &DomainError{Code: ErrCodeAlreadyExists, Message: fmt.Sprintf("%s already exists", resource)}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{Code: ErrCodeConflict, Message: message}
}

func NewValidationError(message string) *DomainError {
	return &DomainError{Code: ErrCodeInvalidInput, Message: message}
}

func NewVersionConflictError() *DomainError {
	return &DomainError{Code: ErrCodeVersionConflict, Message: "version conflict, resource was modified"}
}

func NewInternalError(message string, err error) *DomainError {
	return &DomainError{Code: ErrCodeInternal, Message: message, Err: err}
}

func NewDatabaseError(err error) *DomainError {
	return &DomainError{Code: ErrCodeDatabaseError, Message: "database operation failed", Err: err}
}

func NewExternalServiceError(service string, err error) *DomainError {
	return &DomainError{Code: ErrCodeExternalService, Message: fmt.Sprintf("%s service error", service), Err: err}
}

func NewInvalidOAuthStateError() *DomainError {
	return &DomainError{Code: ErrCodeInvalidOAuthState, Message: "invalid or expired OAuth state"}
}


