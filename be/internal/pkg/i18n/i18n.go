package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
)

const (
	translationsDir = "translations"
	translationsExt = ".json"
	DefaultLanguage = "en"
)

//go:embed translations/*.json
var translationsFS embed.FS

var service *Service

type Service struct {
	translations map[string]map[string]any // lang -> nested map
	supported    []string
}

func init() {
	service = &Service{
		translations: make(map[string]map[string]any),
		supported:    []string{},
	}
	if err := service.loadTranslations(); err != nil {
		panic(fmt.Errorf("load translations: %w", err))
	}
	slog.Info("i18n initialized", "supported_languages", service.supported)
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
			return fmt.Errorf("load translation file %q: %w", filename, err)
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

	if !s.isSupported(DefaultLanguage) {
		return fmt.Errorf("default language %q is not supported", DefaultLanguage)
	}

	return nil
}

func Translate(key, lang string, params map[string]string) string {
	return service.translate(key, lang, params)
}

func IsSupported(lang string) bool {
	return service.isSupported(lang)
}

func NormalizeLanguage(lang string) string {
	return service.normalizeLanguage(lang)
}

func GetSupportedLanguages() []string {
	return service.supported
}

// GetLanguageName returns the native display name for a language code
// It looks up the name from the language's own translation file (e.g., "English" for "en", "العربية" for "ar")
func GetLanguageName(langCode, uiLang string) string {

	// Look up language.name.{langCode} from the language's own translation file
	// This ensures we always show native names (e.g., "English", "العربية", "Русский")
	key := fmt.Sprintf("language.name.%s", langCode)
	name := service.translate(key, langCode, nil)

	// If translation not found, fallback to the code itself
	if name == key {
		return langCode
	}
	return name
}

func (s *Service) isSupported(lang string) bool {
	return slices.Contains(s.supported, lang)
}

// normalizeLanguage: "ar-SA" -> "ar", "en-US" -> "en", "unknown" -> "en" (fallback)
func (s *Service) normalizeLanguage(lang string) string {
	if lang == "" {
		return DefaultLanguage
	}

	parts := strings.Split(lang, "-")
	baseLang := strings.ToLower(parts[0])

	if s.isSupported(baseLang) {
		return baseLang
	}

	return DefaultLanguage
}

func (s *Service) translate(key, lang string, params map[string]string) string {
	lang = s.normalizeLanguage(lang)

	if msg := s.lookup(key, lang); msg != "" {
		return s.interpolate(msg, params)
	}

	if lang != DefaultLanguage {
		if msg := s.lookup(key, DefaultLanguage); msg != "" {
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
