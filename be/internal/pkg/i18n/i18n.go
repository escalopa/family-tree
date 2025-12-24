package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const (
	translationsDir = "translations"
	translationsExt = ".json"
	fallbackLang    = "en"
)

//go:embed translations/*.json
var translationsFS embed.FS

var (
	service     *Service
	serviceOnce sync.Once
)

type Service struct {
	translations map[string]map[string]any // lang -> nested map
	supported    []string
	fallback     string
}

func Init() error {
	var err error
	serviceOnce.Do(func() {
		service = &Service{
			translations: make(map[string]map[string]any),
			supported:    []string{},
			fallback:     fallbackLang,
		}
		err = service.loadTranslations()
	})
	return err
}

func (s *Service) loadTranslations() error {
	entries, err := translationsFS.ReadDir(translationsDir)
	if err != nil {
		return fmt.Errorf("read translations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := filepath.Ext(filename)

		if ext != translationsExt {
			continue
		}

		langCode := strings.TrimSuffix(filename, ext)
		s.supported = append(s.supported, langCode)

		data, err := translationsFS.ReadFile(filepath.Join(translationsDir, filename))
		if err != nil {
			return fmt.Errorf("load translation file %s: %w", filename, err)
		}

		var translations map[string]any
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("parse translation file %s: %w", filename, err)
		}

		s.translations[langCode] = translations
	}

	if len(s.supported) == 0 {
		return fmt.Errorf("no translation files found")
	}

	// If service.fallback doesn't exist, use first available language as fallback
	if !s.isSupported(s.fallback) {
		s.fallback = s.supported[0]
	}

	return nil
}

func Translate(key, lang string, params map[string]string) string {
	checkServiceInitialized()
	return service.translate(key, lang, params)
}

func IsSupported(lang string) bool {
	checkServiceInitialized()
	return service.isSupported(lang)
}

func NormalizeLanguage(lang string) string {
	checkServiceInitialized()
	return service.normalizeLanguage(lang)
}

func GetSupportedLanguages() []string {
	checkServiceInitialized()
	return service.supported
}

func checkServiceInitialized() {
	if service == nil {
		panic("i18n service not initialized")
	}
}

func (s *Service) isSupported(lang string) bool {
	return slices.Contains(s.supported, lang)
}

// normalizeLanguage: "ar-SA" -> "ar", "en-US" -> "en", "unknown" -> "en" (fallback)
func (s *Service) normalizeLanguage(lang string) string {
	if lang == "" {
		return s.fallback
	}

	parts := strings.Split(lang, "-")
	baseLang := strings.ToLower(parts[0])

	if s.isSupported(baseLang) {
		return baseLang
	}

	return s.fallback
}

func (s *Service) translate(key, lang string, params map[string]string) string {
	lang = s.normalizeLanguage(lang)

	if msg := s.lookup(key, lang); msg != "" {
		return s.interpolate(msg, params)
	}

	if lang != s.fallback {
		if msg := s.lookup(key, s.fallback); msg != "" {
			return s.interpolate(msg, params)
		}
	}

	return key
}

func (s *Service) lookup(key, lang string) string {
	translations, exists := s.translations[lang]
	if !exists {
		return ""
	}

	// "error.not_found" -> ["error", "not_found"]
	parts := strings.Split(key, ".")

	var current any = translations
	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return ""
}

func (s *Service) interpolate(template string, params map[string]string) string {
	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
