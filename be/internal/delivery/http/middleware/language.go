package middleware

import (
	"github.com/escalopa/family-tree/internal/pkg/i18n"
	"github.com/gin-gonic/gin"
)

const (
	keyInterfaceLanguage = "interface_language"
	headerAcceptLanguage = "Accept-Language"
)

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader(headerAcceptLanguage)

		lang = i18n.NormalizeLanguage(lang)

		c.Set(keyInterfaceLanguage, lang)
		c.Next()
	}
}

func GetInterfaceLanguage(c *gin.Context) string {
	lang, exists := c.Get(keyInterfaceLanguage)
	if !exists {
		return "en"
	}
	return lang.(string)
}
