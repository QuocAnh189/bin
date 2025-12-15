package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the root server configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	JWT      JWTConfig      `json:"jwt"`
	Session  SessionConfig  `json:"session"`
	Registry RegistryConfig `json:"registry"`
	Storage  StorageConfig  `json:"storage"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Addr         string    `json:"addr"`
	ReadTimeout  int       `json:"read_timeout"`
	WriteTimeout int       `json:"write_timeout"`
	IdleTimeout  int       `json:"idle_timeout"`
	TLS          TLSConfig `json:"tls"`
	CORS         CORSConfig `json:"cors"`
}

// TLSConfig holds TLS settings
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret          string `json:"secret"`
	AccessTokenTTL  int    `json:"access_token_ttl"`  // minutes
	RefreshTokenTTL int    `json:"refresh_token_ttl"` // hours
}

// SessionConfig holds session management settings
type SessionConfig struct {
	DefaultTTL    int `json:"default_ttl"`    // minutes
	CleanupPeriod int `json:"cleanup_period"` // minutes
}

// RegistryConfig holds service registry settings
type RegistryConfig struct {
	HealthCheckInterval int `json:"health_check_interval"` // seconds
	HealthCheckTimeout  int `json:"health_check_timeout"`  // seconds
}

// StorageConfig holds storage backend settings
type StorageConfig struct {
	Type     string         `json:"type"` // redis, postgres, memory
	Redis    RedisConfig    `json:"redis"`
	Postgres PostgresConfig `json:"postgres"`
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// PostgresConfig holds PostgreSQL connection settings
type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// LogConfig holds logging settings
type LogConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
}

// Load loads configuration from environment and files
func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/development/config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Override with environment variables
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		cfg.Storage.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Storage.Redis.Password = password
	}

	return &cfg, nil
}

// ConfigRepository defines the interface for configuration storage
type ConfigRepository interface {
	Get(serviceID, version string) (map[string]any, error)
	Set(serviceID, version string, config map[string]any) error
	Delete(serviceID, version string) error
	List(serviceID string) ([]string, error)
}
