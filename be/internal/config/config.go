package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	OAuth    OAuthConfig    `yaml:"oauth"`
	JWT      JWTConfig      `yaml:"jwt"`
	S3       S3Config       `yaml:"s3"`
}

type ServerConfig struct {
	Port           string       `yaml:"port" env:"SERVER_PORT"`
	Mode           string       `yaml:"mode" env:"GIN_MODE"`
	LogLevel       string       `yaml:"log_level" env:"LOG_LEVEL"`
	AllowedOrigins []string     `yaml:"allowed_origins"`
	Cookie         CookieConfig `yaml:"cookie"`
}

type CookieConfig struct {
	AccessTokenMaxAge  int    `yaml:"access_token_max_age" env:"COOKIE_ACCESS_TOKEN_MAX_AGE"`
	RefreshTokenMaxAge int    `yaml:"refresh_token_max_age" env:"COOKIE_REFRESH_TOKEN_MAX_AGE"`
	SessionIDMaxAge    int    `yaml:"session_id_max_age" env:"COOKIE_SESSION_ID_MAX_AGE"`
	Path               string `yaml:"path" env:"COOKIE_PATH"`
	Domain             string `yaml:"domain" env:"COOKIE_DOMAIN"`
	Secure             bool   `yaml:"secure" env:"COOKIE_SECURE"`
	HttpOnly           bool   `yaml:"http_only" env:"COOKIE_HTTP_ONLY"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     string `yaml:"port" env:"DB_PORT"`
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	Name     string `yaml:"name" env:"DB_NAME"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE"`
}

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `yaml:"providers"`
	RedirectBaseURL string                         `yaml:"redirect_base_url" env:"OAUTH_REDIRECT_BASE_URL"`
}

type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id" env:"CLIENT_ID"`
	ClientSecret string   `yaml:"client_secret" env:"CLIENT_SECRET"`
	Scopes       []string `yaml:"scopes"`
	UserInfoURL  string   `yaml:"user_info_url" env:"USER_INFO_URL"`
}

type JWTConfig struct {
	Secret        string        `yaml:"secret" env:"JWT_SECRET"`
	AccessExpiry  time.Duration `yaml:"access_expiry" env:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry" env:"JWT_REFRESH_EXPIRY"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint" env:"S3_ENDPOINT"`
	Region    string `yaml:"region" env:"S3_REGION"`
	Bucket    string `yaml:"bucket" env:"S3_BUCKET"`
	AccessKey string `yaml:"access_key" env:"S3_ACCESS_KEY"`
	SecretKey string `yaml:"secret_key" env:"S3_SECRET_KEY"`
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

	viper.AutomaticEnv()
	bindEnvFromTags(reflect.TypeOf(Config{}), "")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func bindEnvFromTags(t reflect.Type, prefix string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		yamlTag := field.Tag.Get("yaml")
		envTag := field.Tag.Get("env")

		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		yamlKey := strings.Split(yamlTag, ",")[0]
		fullKey := yamlKey
		if prefix != "" {
			fullKey = prefix + "." + yamlKey
		}

		if envTag != "" {
			viper.BindEnv(fullKey, envTag)
		}

		if field.Type.Kind() == reflect.Struct {
			bindEnvFromTags(field.Type, fullKey)
		}
	}
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
