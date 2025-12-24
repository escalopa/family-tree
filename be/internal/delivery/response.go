package delivery

import (
	"errors"
	"net/http"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/i18n"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const keyUILanguage = "ui_language"

func getUILanguage(c *gin.Context) string {
	lang, _ := c.Get(keyUILanguage)
	return lang.(string)
}

func Error(c *gin.Context, err error) {
	uiLang := getUILanguage(c)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		message := i18n.TranslateValidationErrors(validationErrs, uiLang)
		c.JSON(http.StatusBadRequest, dto.Response{
			Success:   false,
			Error:     message,
			ErrorCode: domain.ErrCodeInvalidInput.String(),
		})
		return
	}

	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		translatedMsg := i18n.Translate(
			domainErr.TranslationKey,
			uiLang,
			domainErr.Params,
		)

		c.JSON(domainErr.HTTPStatusCode(), dto.Response{
			Success:   false,
			Error:     translatedMsg,
			ErrorCode: domainErr.Code.String(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, dto.Response{
		Success:   false,
		Error:     i18n.Translate("error.internal", uiLang, nil),
		ErrorCode: domain.ErrCodeInternal.String(),
	})
}

func Success(c *gin.Context, translationKey string, params map[string]string) {
	uiLang := getUILanguage(c)
	message := i18n.Translate(translationKey, uiLang, params)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    message,
	})
}

func SuccessWithData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    data,
	})
}
