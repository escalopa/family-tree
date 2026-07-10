package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server" json:"server"`
	Database    DatabaseConfig    `mapstructure:"database" json:"database"`
	Redis       RedisConfig       `mapstructure:"redis" json:"redis"`
	OAuth       OAuthConfig       `mapstructure:"oauth" json:"oauth"`
	JWT         JWTConfig         `mapstructure:"jwt" json:"jwt"`
	S3          S3Config          `mapstructure:"s3" json:"s3"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit" json:"rate_limit"`
	Upload      UploadConfig      `mapstructure:"upload" json:"upload"`
	Maintenance MaintenanceConfig `mapstructure:"maintenance" json:"maintenance"`
}

type ServerConfig struct {
	Port           string       `mapstructure:"port" env:"SERVER_PORT" json:"port"`
	Mode           string       `mapstructure:"mode" env:"GIN_MODE"`
	LogLevel       string       `mapstructure:"log_level" env:"LOG_LEVEL" json:"log_level"`
	AllowedOrigins []string     `mapstructure:"allowed_origins" env:"ALLOWED_ORIGINS" json:"allowed_origins"`
	EnableHSTS     bool         `mapstructure:"enable_hsts" env:"ENABLE_HSTS" json:"enable_hsts"`
	Cookie         CookieConfig `mapstructure:"cookie" json:"cookie"`
}

type CookieConfig struct {
	AccessTokenMaxAge  int    `mapstructure:"access_token_max_age" env:"COOKIE_ACCESS_TOKEN_MAX_AGE" json:"access_token_max_age"`
	RefreshTokenMaxAge int    `mapstructure:"refresh_token_max_age" env:"COOKIE_REFRESH_TOKEN_MAX_AGE" json:"refresh_token_max_age"`
	SessionIDMaxAge    int    `mapstructure:"session_id_max_age" env:"COOKIE_SESSION_ID_MAX_AGE" json:"session_id_max_age"`
	Path               string `mapstructure:"path" env:"COOKIE_PATH"`
	Domain             string `mapstructure:"domain" env:"COOKIE_DOMAIN" json:"domain"`
	Secure             bool   `mapstructure:"secure" env:"COOKIE_SECURE" json:"secure"`
	HttpOnly           bool   `mapstructure:"http_only" env:"COOKIE_HTTP_ONLY" json:"http_only"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn" env:"DATABASE_DSN" json:"dsn"`
}

type RedisConfig struct {
	URI string `mapstructure:"uri" env:"REDIS_URI" json:"uri"`
}

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `mapstructure:"providers" json:"providers"`
	RedirectBaseURL string                         `mapstructure:"redirect_base_url" env:"OAUTH_REDIRECT_BASE_URL" json:"redirect_base_url"`
	providerOrder   []string                       // cached computed order
}

type OAuthProviderConfig struct {
	Order        int      `mapstructure:"order" json:"order"`
	ClientID     string   `mapstructure:"client_id" env:"CLIENT_ID" json:"client_id"`
	ClientSecret string   `mapstructure:"client_secret" env:"CLIENT_SECRET" json:"client_secret"`
	AuthURL      string   `mapstructure:"auth_url" env:"AUTH_URL" json:"auth_url"`
	TokenURL     string   `mapstructure:"token_url" env:"TOKEN_URL" json:"token_url"`
	Scopes       []string `mapstructure:"scopes" json:"scopes"`
	UserInfoURL  string   `mapstructure:"user_info_url" env:"USER_INFO_URL" json:"user_info_url"`
	IDField      string   `mapstructure:"id_field" env:"ID_FIELD" json:"id_field"`
	EmailField   string   `mapstructure:"email_field" env:"EMAIL_FIELD" json:"email_field"`
	NameField    string   `mapstructure:"name_field" env:"NAME_FIELD" json:"name_field"`
	PictureField string   `mapstructure:"picture_field" env:"PICTURE_FIELD" json:"picture_field"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret" env:"JWT_SECRET" json:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry" env:"JWT_ACCESS_EXPIRY" json:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry" env:"JWT_REFRESH_EXPIRY" json:"refresh_expiry"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint" env:"S3_ENDPOINT" json:"endpoint"`
	Region    string `mapstructure:"region" env:"S3_REGION" json:"region"`
	Bucket    string `mapstructure:"bucket" env:"S3_BUCKET" json:"bucket"`
	AccessKey string `mapstructure:"access_key" env:"S3_ACCESS_KEY" json:"access_key"`
	SecretKey string `mapstructure:"secret_key" env:"S3_SECRET_KEY" json:"secret_key"`
}

type RateLimitRule struct {
	Enabled  bool          `mapstructure:"enabled" json:"enabled"`
	Requests int           `mapstructure:"requests" json:"requests"`
	Window   time.Duration `mapstructure:"window" json:"window"`
	Prefix   string        `mapstructure:"prefix" json:"prefix"`
}

type RateLimitConfig struct {
	Auth   RateLimitRule `mapstructure:"auth" json:"auth"`
	API    RateLimitRule `mapstructure:"api" json:"api"`
	Upload RateLimitRule `mapstructure:"upload" json:"upload"`
}

type UploadConfig struct {
	MaxImageSize     int64    `mapstructure:"max_image_size" env:"UPLOAD_MAX_IMAGE_SIZE" json:"max_image_size"` // bytes
	AllowedImageExts []string `mapstructure:"allowed_image_extensions" env:"UPLOAD_ALLOWED_IMAGE_EXTENSIONS" json:"allowed_image_extensions"`
}

type MaintenanceConfig struct {
	CleanupInterval time.Duration `mapstructure:"cleanup_interval" env:"MAINTENANCE_CLEANUP_INTERVAL" json:"cleanup_interval"`
}

func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	length := len([]rune(secret))
	if length <= 8 {
		return "****"
	}
	return string([]rune(secret)[:4]) + "****" + string([]rune(secret)[length-4:])
}

func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	// Handle key=value format: "host=localhost password=secret dbname=test"
	if strings.Contains(dsn, "password=") {
		parts := strings.Split(dsn, " ")
		for i, part := range parts {
			if password, found := strings.CutPrefix(part, "password="); found {
				parts[i] = "password=" + maskSecret(password)
			}
		}
		return strings.Join(parts, " ")
	}

	// Handle URI format: "postgresql://user:password@host:port/dbname"
	if strings.Contains(dsn, "://") && strings.Contains(dsn, ":") && strings.Contains(dsn, "@") {
		// Find password between "://" and "@"
		schemeEnd := strings.Index(dsn, "://")
		atIndex := strings.Index(dsn, "@")
		if schemeEnd != -1 && atIndex != -1 {
			beforeAuth := dsn[:schemeEnd+3]
			afterHost := dsn[atIndex:]
			authPart := dsn[schemeEnd+3 : atIndex]

			if colonIndex := strings.Index(authPart, ":"); colonIndex != -1 {
				username := authPart[:colonIndex]
				password := authPart[colonIndex+1:]
				return beforeAuth + username + ":" + maskSecret(password) + afterHost
			}
		}
	}

	return dsn
}

func (c *Config) maskedConfig() Config {
	masked := *c
	masked.Database.DSN = maskDSN(c.Database.DSN)
	masked.JWT.Secret = maskSecret(c.JWT.Secret)
	masked.OAuth.Providers = make(map[string]OAuthProviderConfig)
	for provider, config := range c.OAuth.Providers {
		maskedProvider := config
		maskedProvider.ClientID = maskSecret(config.ClientID)
		maskedProvider.ClientSecret = maskSecret(config.ClientSecret)
		masked.OAuth.Providers[provider] = maskedProvider
	}
	masked.S3.AccessKey = maskSecret(c.S3.AccessKey)
	masked.S3.SecretKey = maskSecret(c.S3.SecretKey)
	return masked
}

func (c *Config) String() string {
	masked := c.maskedConfig()
	bytes, _ := json.Marshal(masked)
	return string(bytes)
}

const (
	envConfigPath     = "CONFIG_PATH"
	defaultConfigPath = "."
)

func Load() (*Config, error) {
	configPath := os.Getenv(envConfigPath)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	setDefaults()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) || strings.TrimSpace(os.Getenv("DATABASE_DSN")) == "" {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		slog.Info("Config.Load: config file not found; using environment variables")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.applyEnvOverrides()
	cfg.OAuth.computeProviderOrder()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	slog.Info("Config", "config", cfg.String())

	return &cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Database.DSN) == "" {
		return fmt.Errorf("DATABASE_DSN is required")
	}
	if strings.TrimSpace(c.JWT.Secret) == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.OAuth.Providers) > 0 && strings.TrimSpace(c.OAuth.RedirectBaseURL) == "" {
		return fmt.Errorf("OAUTH_REDIRECT_BASE_URL is required when OAuth providers are enabled")
	}
	return nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("server.cookie.access_token_max_age", 3600)
	viper.SetDefault("server.cookie.refresh_token_max_age", 604800)
	viper.SetDefault("server.cookie.session_id_max_age", 604800)
	viper.SetDefault("server.cookie.path", "/")
	viper.SetDefault("server.cookie.secure", true)
	viper.SetDefault("server.cookie.http_only", true)
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("upload.max_image_size", 3145728)
	viper.SetDefault("upload.allowed_image_extensions", []string{".jpg", ".jpeg", ".png", ".gif", ".webp"})
	viper.SetDefault("maintenance.cleanup_interval", "1h")
}

func (c *Config) applyEnvOverrides() {
	if port := firstNonEmptyEnv("PORT", "SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if mode := firstNonEmptyEnv("GIN_MODE", "SERVER_MODE"); mode != "" {
		c.Server.Mode = mode
	}
	if logLevel := firstNonEmptyEnv("LOG_LEVEL", "SERVER_LOG_LEVEL"); logLevel != "" {
		c.Server.LogLevel = logLevel
	}
	if origins := splitCSV(os.Getenv("ALLOWED_ORIGINS")); len(origins) > 0 {
		c.Server.AllowedOrigins = origins
	}
	if value, ok := boolEnv("ENABLE_HSTS"); ok {
		c.Server.EnableHSTS = value
	}
	if value, ok := boolEnv("COOKIE_SECURE"); ok {
		c.Server.Cookie.Secure = value
	}
	if value, ok := boolEnv("COOKIE_HTTP_ONLY"); ok {
		c.Server.Cookie.HttpOnly = value
	}
	if domain := os.Getenv("COOKIE_DOMAIN"); domain != "" {
		c.Server.Cookie.Domain = domain
	}
	if path := os.Getenv("COOKIE_PATH"); path != "" {
		c.Server.Cookie.Path = path
	}
	if value, ok := intEnv("COOKIE_ACCESS_TOKEN_MAX_AGE"); ok {
		c.Server.Cookie.AccessTokenMaxAge = value
	}
	if value, ok := intEnv("COOKIE_REFRESH_TOKEN_MAX_AGE"); ok {
		c.Server.Cookie.RefreshTokenMaxAge = value
	}
	if value, ok := intEnv("COOKIE_SESSION_ID_MAX_AGE"); ok {
		c.Server.Cookie.SessionIDMaxAge = value
	}
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		c.Database.DSN = dsn
	}
	if uri := os.Getenv("REDIS_URI"); uri != "" {
		c.Redis.URI = uri
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.JWT.Secret = secret
	}
	if value, ok := durationEnv("JWT_ACCESS_EXPIRY"); ok {
		c.JWT.AccessExpiry = value
	}
	if value, ok := durationEnv("JWT_REFRESH_EXPIRY"); ok {
		c.JWT.RefreshExpiry = value
	}
	if value, ok := durationEnv("MAINTENANCE_CLEANUP_INTERVAL"); ok {
		c.Maintenance.CleanupInterval = value
	}

	c.applyS3Env()
	c.applyRateLimitEnv()
	c.applyOAuthEnv()
}

func (c *Config) applyS3Env() {
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		c.S3.Endpoint = endpoint
	}
	if region := os.Getenv("S3_REGION"); region != "" {
		c.S3.Region = region
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		c.S3.Bucket = bucket
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		c.S3.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		c.S3.SecretKey = secretKey
	}
	if value, ok := int64Env("UPLOAD_MAX_IMAGE_SIZE"); ok {
		c.Upload.MaxImageSize = value
	}
	if extensions := splitCSV(os.Getenv("UPLOAD_ALLOWED_IMAGE_EXTENSIONS")); len(extensions) > 0 {
		c.Upload.AllowedImageExts = extensions
	}
}

func (c *Config) applyRateLimitEnv() {
	if c.Redis.URI == "" {
		c.RateLimit.Auth.Enabled = false
		c.RateLimit.API.Enabled = false
		c.RateLimit.Upload.Enabled = false
		return
	}

	applyRateLimitRuleEnv("AUTH", &c.RateLimit.Auth)
	applyRateLimitRuleEnv("API", &c.RateLimit.API)
	applyRateLimitRuleEnv("UPLOAD", &c.RateLimit.Upload)
}

func applyRateLimitRuleEnv(prefix string, rule *RateLimitRule) {
	if value, ok := boolEnv("RATE_LIMIT_" + prefix + "_ENABLED"); ok {
		rule.Enabled = value
	}
	if value, ok := intEnv("RATE_LIMIT_" + prefix + "_REQUESTS"); ok {
		rule.Requests = value
	}
	if value, ok := durationEnv("RATE_LIMIT_" + prefix + "_WINDOW"); ok {
		rule.Window = value
	}
	if value := os.Getenv("RATE_LIMIT_" + prefix + "_PREFIX"); value != "" {
		rule.Prefix = value
	}
}

func (c *Config) applyOAuthEnv() {
	if redirectBaseURL := os.Getenv("OAUTH_REDIRECT_BASE_URL"); redirectBaseURL != "" {
		c.OAuth.RedirectBaseURL = strings.TrimRight(redirectBaseURL, "/")
	}

	enabledProviders := splitCSV(os.Getenv("OAUTH_ENABLED_PROVIDERS"))
	if len(enabledProviders) == 0 {
		enabledProviders = slices.Collect(maps.Keys(c.OAuth.Providers))
	}
	if len(enabledProviders) == 0 && strings.EqualFold(os.Getenv("ENABLE_MOCK_AUTH"), "true") {
		enabledProviders = []string{"mock"}
	}
	if len(enabledProviders) == 0 {
		return
	}

	providers := make(map[string]OAuthProviderConfig, len(enabledProviders))
	for index, provider := range enabledProviders {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" {
			continue
		}

		providerCfg := c.OAuth.Providers[provider]
		if providerCfg.Order == 0 {
			providerCfg.Order = index + 1
		}

		envPrefix := "OAUTH_PROVIDER_" + strings.ToUpper(strings.ReplaceAll(provider, "-", "_")) + "_"
		if value := os.Getenv(envPrefix + "CLIENT_ID"); value != "" {
			providerCfg.ClientID = value
		}
		if value := os.Getenv(envPrefix + "CLIENT_SECRET"); value != "" {
			providerCfg.ClientSecret = value
		}
		if value := os.Getenv(envPrefix + "AUTH_URL"); value != "" {
			providerCfg.AuthURL = value
		}
		if value := os.Getenv(envPrefix + "TOKEN_URL"); value != "" {
			providerCfg.TokenURL = value
		}
		if value := os.Getenv(envPrefix + "USER_INFO_URL"); value != "" {
			providerCfg.UserInfoURL = value
		}
		if value := os.Getenv(envPrefix + "ID_FIELD"); value != "" {
			providerCfg.IDField = value
		}
		if value := os.Getenv(envPrefix + "EMAIL_FIELD"); value != "" {
			providerCfg.EmailField = value
		}
		if value := os.Getenv(envPrefix + "NAME_FIELD"); value != "" {
			providerCfg.NameField = value
		}
		if value := os.Getenv(envPrefix + "PICTURE_FIELD"); value != "" {
			providerCfg.PictureField = value
		}
		if scopes := splitCSV(os.Getenv(envPrefix + "SCOPES")); len(scopes) > 0 {
			providerCfg.Scopes = scopes
		}
		if value, ok := intEnv(envPrefix + "ORDER"); ok {
			providerCfg.Order = value
		}

		providers[provider] = providerCfg
	}

	c.OAuth.Providers = providers
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func boolEnv(key string) (bool, bool) {
	raw := os.Getenv(key)
	if raw == "" {
		return false, false
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("Config: invalid bool env; ignoring", "key", key, "value", raw)
		return false, false
	}
	return value, true
}

func intEnv(key string) (int, bool) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("Config: invalid int env; ignoring", "key", key, "value", raw)
		return 0, false
	}
	return value, true
}

func int64Env(key string) (int64, bool) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, false
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		slog.Warn("Config: invalid int64 env; ignoring", "key", key, "value", raw)
		return 0, false
	}
	return value, true
}

func durationEnv(key string) (time.Duration, bool) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, false
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		slog.Warn("Config: invalid duration env; ignoring", "key", key, "value", raw)
		return 0, false
	}
	return value, true
}

func (c *DatabaseConfig) ConnectionString() string {
	return c.DSN
}

func (c *JWTConfig) ParseAccessExpiry() time.Duration {
	return c.AccessExpiry
}

func (c *JWTConfig) ParseRefreshExpiry() time.Duration {
	return c.RefreshExpiry
}

func (c *OAuthConfig) GetRedirectURL(provider string) string {
	return fmt.Sprintf("%s/auth/%s/callback", c.RedirectBaseURL, provider)
}

func (c *OAuthConfig) computeProviderOrder() {
	c.providerOrder = slices.Collect(maps.Keys(c.Providers))

	slices.SortFunc(c.providerOrder, func(a, b string) int {
		orderA := c.Providers[a].Order
		if orderA == 0 {
			orderA = math.MaxInt
		}
		orderB := c.Providers[b].Order
		if orderB == 0 {
			orderB = math.MaxInt
		}

		if orderA == orderB {
			return strings.Compare(a, b)
		}

		return orderA - orderB
	})
}

func (c *OAuthConfig) GetProviderOrder() []string {
	return c.providerOrder
}
