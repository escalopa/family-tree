package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	S3       S3Config       `mapstructure:"s3"`
}

type ServerConfig struct {
	Port           string       `mapstructure:"port" env:"SERVER_PORT"`
	Mode           string       `mapstructure:"mode" env:"GIN_MODE"`
	LogLevel       string       `mapstructure:"log_level" env:"LOG_LEVEL"`
	AllowedOrigins []string     `mapstructure:"allowed_origins" env:"ALLOWED_ORIGINS"`
	Cookie         CookieConfig `mapstructure:"cookie"`
}

type CookieConfig struct {
	AccessTokenMaxAge  int    `mapstructure:"access_token_max_age" env:"COOKIE_ACCESS_TOKEN_MAX_AGE"`
	RefreshTokenMaxAge int    `mapstructure:"refresh_token_max_age" env:"COOKIE_REFRESH_TOKEN_MAX_AGE"`
	SessionIDMaxAge    int    `mapstructure:"session_id_max_age" env:"COOKIE_SESSION_ID_MAX_AGE"`
	Path               string `mapstructure:"path" env:"COOKIE_PATH"`
	Domain             string `mapstructure:"domain" env:"COOKIE_DOMAIN"`
	Secure             bool   `mapstructure:"secure" env:"COOKIE_SECURE"`
	HttpOnly           bool   `mapstructure:"http_only" env:"COOKIE_HTTP_ONLY"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" env:"DB_HOST"`
	Port     string `mapstructure:"port" env:"DB_PORT"`
	User     string `mapstructure:"user" env:"DB_USER"`
	Password string `mapstructure:"password" env:"DB_PASSWORD"`
	Name     string `mapstructure:"name" env:"DB_NAME"`
	SSLMode  string `mapstructure:"sslmode" env:"DB_SSLMODE"`
}

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `mapstructure:"providers"`
	RedirectBaseURL string                         `mapstructure:"redirect_base_url" env:"OAUTH_REDIRECT_BASE_URL"`
}

type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id" env:"CLIENT_ID"`
	ClientSecret string   `mapstructure:"client_secret" env:"CLIENT_SECRET"`
	Scopes       []string `mapstructure:"scopes"`
	UserInfoURL  string   `mapstructure:"user_info_url" env:"USER_INFO_URL"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret" env:"JWT_SECRET"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry" env:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry" env:"JWT_REFRESH_EXPIRY"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint" env:"S3_ENDPOINT"`
	Region    string `mapstructure:"region" env:"S3_REGION"`
	Bucket    string `mapstructure:"bucket" env:"S3_BUCKET"`
	AccessKey string `mapstructure:"access_key" env:"S3_ACCESS_KEY"`
	SecretKey string `mapstructure:"secret_key" env:"S3_SECRET_KEY"`
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

	slog.Info("Config", "cfg", cfg)

	return &cfg, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
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
