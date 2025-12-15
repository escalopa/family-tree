package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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
	Host     string `mapstructure:"host" env:"DB_HOST" json:"host"`
	Port     string `mapstructure:"port" env:"DB_PORT" json:"port"`
	User     string `mapstructure:"user" env:"DB_USER" json:"user"`
	Password string `mapstructure:"password" env:"DB_PASSWORD" json:"password"`
	Name     string `mapstructure:"name" env:"DB_NAME" json:"name"`
	SSLMode  string `mapstructure:"sslmode" env:"DB_SSLMODE" json:"sslmode"`
}

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `mapstructure:"providers" json:"providers"`
	RedirectBaseURL string                         `mapstructure:"redirect_base_url" env:"OAUTH_REDIRECT_BASE_URL" json:"redirect_base_url"`
}

type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id" env:"CLIENT_ID" json:"-"`
	ClientSecret string   `mapstructure:"client_secret" env:"CLIENT_SECRET" json:"-"`
	Scopes       []string `mapstructure:"scopes" json:"scopes"`
	UserInfoURL  string   `mapstructure:"user_info_url" env:"USER_INFO_URL" json:"user_info_url"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret" env:"JWT_SECRET" json:"-"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry" env:"JWT_ACCESS_EXPIRY" json:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry" env:"JWT_REFRESH_EXPIRY" json:"refresh_expiry"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint" env:"S3_ENDPOINT" json:"endpoint"`
	Region    string `mapstructure:"region" env:"S3_REGION" json:"region"`
	Bucket    string `mapstructure:"bucket" env:"S3_BUCKET" json:"bucket"`
	AccessKey string `mapstructure:"access_key" env:"S3_ACCESS_KEY" json:"-"`
	SecretKey string `mapstructure:"secret_key" env:"S3_SECRET_KEY" json:"-"`
}

func (c *Config) String() string {
	bytes, _ := json.Marshal(c)
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

	slog.Info("Config", "config", cfg.String())

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
