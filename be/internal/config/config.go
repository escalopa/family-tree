package config

import (
	"fmt"
	"os"
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
	Port     string `yaml:"port"`
	Mode     string `yaml:"mode"`
	LogLevel string `yaml:"log_level"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type OAuthConfig struct {
	Providers       map[string]OAuthProviderConfig `yaml:"providers"`
	RedirectBaseURL string                         `yaml:"redirect_base_url"`
}

type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	Scopes       []string `yaml:"scopes"`
}

type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

func Load() (*Config, error) {
	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "." // Default to current directory
	}

	// Set config name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	// Enable automatic environment variable override
	viper.AutomaticEnv()

	// Bind environment variables to config keys
	bindEnvVariables()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func bindEnvVariables() {
	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "GIN_MODE")
	viper.BindEnv("server.log_level", "LOG_LEVEL")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.access_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")

	// OAuth
	viper.BindEnv("oauth.providers.google.client_id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.providers.google.client_secret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.redirect_base_url", "OAUTH_REDIRECT_BASE_URL")

	// S3
	viper.BindEnv("s3.endpoint", "S3_ENDPOINT")
	viper.BindEnv("s3.region", "S3_REGION")
	viper.BindEnv("s3.bucket", "S3_BUCKET")
	viper.BindEnv("s3.access_key", "S3_ACCESS_KEY")
	viper.BindEnv("s3.secret_key", "S3_SECRET_KEY")
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
