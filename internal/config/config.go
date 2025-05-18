package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	Defaults  DefaultConfig
	Logging   LoggingConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host             string        `mapstructure:"host"`
	Port             int           `mapstructure:"port"`
	User             string        `mapstructure:"user"`
	Password         string        `mapstructure:"password"`
	Name             string        `mapstructure:"name"`
	SSLMode          string        `mapstructure:"ssl_mode"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	MaxOpenConns     int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime  time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationEnabled bool          `mapstructure:"migration_enabled"`
}

type SecurityConfig struct {
	EncryptionKey        string `mapstructure:"encryption_key"`
	EncryptionKeyVersion int    `mapstructure:"encryption_key_version"`
}

type DefaultConfig struct {
	Provider  string           `mapstructure:"provider"`
	Model     string           `mapstructure:"model"`
	Providers []ProviderConfig `mapstructure:"providers"` // Changed from "default_providers"
}

type ProviderConfig struct {
	Name       string        `mapstructure:"name"`
	BaseURL    string        `mapstructure:"base_url"`
	APIKey     string        `mapstructure:"api_key"`
	Default    bool          `mapstructure:"default"`
	Models     []ModelConfig `mapstructure:"models"`
	AuthMethod string        `mapstructure:"auth_method"` // "header" or "query_param"
}

type ModelConfig struct {
	Name         string `mapstructure:"name"`
	Parameters   string `mapstructure:"parameters"`
	Config       string `mapstructure:"config"`
	ResponsePath string `mapstructure:"response_path"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RateLimitConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Requests    float64  `mapstructure:"requests"`
	Window      string   `mapstructure:"window"`
	IPWhitelist []string `mapstructure:"ip_whitelist"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)

	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("database.migration_enabled", true)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.requests", 100)
	v.SetDefault("rate_limit.window", "1m")

	// Bind environment variables to config paths
	_ = v.BindEnv("server.port", "PORT")
	_ = v.BindEnv("server.read_timeout", "READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "IDLE_TIMEOUT")

	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.user", "DB_USER")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.name", "DB_NAME")
	_ = v.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	_ = v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	_ = v.BindEnv("database.migration_enabled", "DB_MIGRATION_ENABLED")

	_ = v.BindEnv("security.encryption_key", "ENCRYPTION_KEY")
	_ = v.BindEnv("security.encryption_key_version", "ENCRYPTION_KEY_VERSION")

	_ = v.BindEnv("defaults.provider", "DEFAULT_PROVIDER")
	_ = v.BindEnv("defaults.model", "DEFAULT_MODEL")

	_ = v.BindEnv("logging.level", "LOG_LEVEL")
	_ = v.BindEnv("logging.format", "LOG_FORMAT")

	_ = v.BindEnv("rate_limit.enabled", "RATE_LIMIT_ENABLED")
	_ = v.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	_ = v.BindEnv("rate_limit.window", "RATE_LIMIT_WINDOW")
	_ = v.BindEnv("rate_limit.ip_whitelist", "RATE_LIMIT_IP_WHITELIST")
	// Configuration sources
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	// Bind environment variables
	_ = v.BindEnv("security.encryption_key", "ENCRYPTION_KEY")
	_ = v.BindEnv("database.password", "DB_PASSWORD")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	for i := range cfg.Defaults.Providers {
		cfg.Defaults.Providers[i].APIKey = os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(cfg.Defaults.Providers[i].Name)))
	}
	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Security.EncryptionKey == "" {
		return fmt.Errorf("encryption key is required")
	}

	// if len(cfg.Security.EncryptionKey) != 32 {
	// 	return fmt.Errorf("encryption key must be 32 bytes")
	// }

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.Defaults.Provider == "" && len(cfg.Defaults.Providers) > 0 {
		cfg.Defaults.Provider = cfg.Defaults.Providers[0].Name
	}

	return nil
}
