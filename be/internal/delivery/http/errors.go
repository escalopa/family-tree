package http

import (
	"errors"
	"net/http"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		c.JSON(domainErr.HTTPStatusCode(), gin.H{
			"success": false,
			"error":   domainErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   "Internal server error",
	})
}
