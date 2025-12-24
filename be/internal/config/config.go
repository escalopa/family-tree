package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	OAuth    OAuthConfig    `mapstructure:"oauth" json:"oauth"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt"`
	S3       S3Config       `mapstructure:"s3" json:"s3"`
}

type ServerConfig struct {
	Port           string       `mapstructure:"port" env:"SERVER_PORT" json:"port"`
	Mode           string       `mapstructure:"mode" env:"GIN_MODE"`
	LogLevel       string       `mapstructure:"log_level" env:"LOG_LEVEL" json:"log_level"`
	AllowedOrigins []string     `mapstructure:"allowed_origins" env:"ALLOWED_ORIGINS" json:"allowed_origins"`
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

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `mapstructure:"providers" json:"providers"`
	RedirectBaseURL string                         `mapstructure:"redirect_base_url" env:"OAUTH_REDIRECT_BASE_URL" json:"redirect_base_url"`
	providerOrder   []string                       // cached computed order
}

type OAuthProviderConfig struct {
	Order        int      `mapstructure:"order" json:"order"`
	ClientID     string   `mapstructure:"client_id" env:"CLIENT_ID" json:"client_id"`
	ClientSecret string   `mapstructure:"client_secret" env:"CLIENT_SECRET" json:"client_secret"`
	Scopes       []string `mapstructure:"scopes" json:"scopes"`
	UserInfoURL  string   `mapstructure:"user_info_url" env:"USER_INFO_URL" json:"user_info_url"`
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

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.OAuth.computeProviderOrder()

	slog.Info("Config", "config", cfg.String())

	return &cfg, nil
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
