package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	OAuth    OAuthConfig    `yaml:"oauth"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Port            string        `yaml:"port"`
	Host            string        `yaml:"host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int32  `yaml:"max_conns"`
	MinConns int32  `yaml:"min_conns"`
}

type OAuthConfig struct {
	Google OAuthProviderConfig `yaml:"google"`
}

type OAuthProviderConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type JWTConfig struct {
	AuthTokenSecret    string        `yaml:"auth_token_secret"`
	RefreshTokenSecret string        `yaml:"refresh_token_secret"`
	AuthTokenExpiry    time.Duration `yaml:"auth_token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
}

func LoadConfig(configName string, configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	if configPath != "" {
		v.AddConfigPath(configPath)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.readtimeout", time.Second*30)
	v.SetDefault("server.writetimeout", time.Second*30)
	v.SetDefault("server.shutdowntimeout", time.Second*10)

	if err := v.ReadInConfig(); err != nil {
		// allow missing file if envs provide values
		var cfg Config
		if err := v.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("unmarshal config without file: %w", err)
		}
		return &cfg, nil
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
