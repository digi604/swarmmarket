package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
}

// Address returns the server address string.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	Host         string        `envconfig:"DB_HOST" default:"localhost"`
	Port         int           `envconfig:"DB_PORT" default:"5432"`
	User         string        `envconfig:"DB_USER" default:"swarmmarket"`
	Password     string        `envconfig:"DB_PASSWORD" default:"swarmmarket"`
	Name         string        `envconfig:"DB_NAME" default:"swarmmarket"`
	SSLMode      string        `envconfig:"DB_SSL_MODE" default:"disable"`
	MaxConns     int           `envconfig:"DB_MAX_CONNS" default:"25"`
	MinConns     int           `envconfig:"DB_MIN_CONNS" default:"5"`
	MaxConnLife  time.Duration `envconfig:"DB_MAX_CONN_LIFE" default:"1h"`
	MaxConnIdle  time.Duration `envconfig:"DB_MAX_CONN_IDLE" default:"30m"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

// Address returns the Redis address string.
func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	APIKeyHeader   string        `envconfig:"AUTH_API_KEY_HEADER" default:"X-API-Key"`
	APIKeyLength   int           `envconfig:"AUTH_API_KEY_LENGTH" default:"32"`
	RateLimitRPS   int           `envconfig:"AUTH_RATE_LIMIT_RPS" default:"100"`
	RateLimitBurst int           `envconfig:"AUTH_RATE_LIMIT_BURST" default:"200"`
	TokenTTL       time.Duration `envconfig:"AUTH_TOKEN_TTL" default:"24h"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

// MustLoad reads configuration and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
