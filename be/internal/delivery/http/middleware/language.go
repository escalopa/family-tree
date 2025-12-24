package middleware

import (
	"github.com/escalopa/family-tree/internal/pkg/i18n"
	"github.com/gin-gonic/gin"
)

const (
	keyUILanguage        = "ui_language"
	headerAcceptLanguage = "Accept-Language"
)

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader(headerAcceptLanguage)

		lang = i18n.NormalizeLanguage(lang)

		c.Set(keyUILanguage, lang)
		c.Next()
	}
}
