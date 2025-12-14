package http

import (
	"errors"
	"net/http"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

// HandleError maps domain errors to HTTP responses
func HandleError(c *gin.Context, err error) {
	statusCode, message := mapDomainError(err)
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

// mapDomainError converts domain errors to HTTP status codes using errors.Is
func mapDomainError(err error) (int, string) {
	// Check for domain errors using errors.Is
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, getErrorMessage(err)
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, getErrorMessage(err)
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, getErrorMessage(err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, getErrorMessage(err)
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, getErrorMessage(err)
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, getErrorMessage(err)
	case errors.Is(err, domain.ErrInvalidOAuthState):
		return http.StatusBadRequest, getErrorMessage(err)
	case errors.Is(err, domain.ErrDatabase):
		return http.StatusInternalServerError, "Database error occurred"
	case errors.Is(err, domain.ErrExternalService):
		return http.StatusServiceUnavailable, "External service error"
	default:
		// Check if it's a DomainError to get the message
		var domainErr *domain.DomainError
		if errors.As(err, &domainErr) {
			return http.StatusInternalServerError, domainErr.Message
		}
		return http.StatusInternalServerError, "Internal server error"
	}
}

func getErrorMessage(err error) string {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Message
	}
	return err.Error()
}


