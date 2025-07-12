package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Swagger   SwaggerConfig   `mapstructure:"swagger"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Environment  string `mapstructure:"environment"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	JWTExpiration int    `mapstructure:"jwt_expiration"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level            string            `mapstructure:"level"`
	Format           string            `mapstructure:"format"`
	StartupLevel     string            `mapstructure:"startup_level"`
	StartupFormat    string            `mapstructure:"startup_format"`
	EnableCaller     bool              `mapstructure:"enable_caller"`
	EnableStacktrace bool              `mapstructure:"enable_stacktrace"`
	File             FileLoggingConfig `mapstructure:"file"`
}

// FileLoggingConfig holds file-based logging configuration
type FileLoggingConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Directory  string `mapstructure:"directory"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // Number of backup files
	MaxAge     int    `mapstructure:"max_age"`     // Days
	Compress   bool   `mapstructure:"compress"`
}

// SwaggerConfig holds Swagger documentation configuration
type SwaggerConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Route       string   `mapstructure:"route"`
	Title       string   `mapstructure:"title"`
	Description string   `mapstructure:"description"`
	Version     string   `mapstructure:"version"`
	Host        string   `mapstructure:"host"`
	BasePath    string   `mapstructure:"base_path"`
	Schemes     []string `mapstructure:"schemes"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Setup viper to read from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Configure viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		// Log the error but continue, as we can still use environment variables and defaults
		fmt.Printf("Warning: %v\n", err)
	}

	// Map environment variables to config fields
	mapEnvVariables()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets essential default values for configuration
// Only critical fallbacks that must be available even without config.yaml
func setDefaults() {
	// Essential server defaults (required for startup)
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.environment", "development")

	// Critical database defaults (required for connection)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")

	// Redis connection defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")

	// Security defaults (critical for auth)
	viper.SetDefault("auth.jwt_expiration", 3600)

	// Minimal logging defaults (only critical fallbacks)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Note: Most configuration is now in config.yaml
	// These defaults are only fallbacks for critical startup requirements
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	// Define possible locations for .env file
	possibleLocations := []string{
		".env",
		filepath.Join(".", ".env"),
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	// Try to load .env file from possible locations
	for _, location := range possibleLocations {
		if _, err := os.Stat(location); err == nil {
			err := godotenv.Load(location)
			if err != nil {
				return fmt.Errorf("error loading .env file from %s: %w", location, err)
			}
			return nil // Successfully loaded
		}
	}

	// If .env file not found, try to load from .env.example
	for _, location := range []string{
		".env.example",
		filepath.Join(".", ".env.example"),
		filepath.Join("..", ".env.example"),
		filepath.Join("..", "..", ".env.example"),
	} {
		if _, err := os.Stat(location); err == nil {
			err := godotenv.Load(location)
			if err != nil {
				return fmt.Errorf("error loading .env.example file from %s: %w", location, err)
			}
			return nil // Successfully loaded
		}
	}

	// It's okay if no .env file is found
	return nil
}

// mapEnvVariables maps environment variables to config fields
// Only maps environment-specific and sensitive configurations
func mapEnvVariables() {
	// Server configuration (environment-specific)
	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("server.host", "HOST")
	_ = viper.BindEnv("server.environment", "ENVIRONMENT")

	// Database configuration (sensitive and environment-specific)
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.dbname", "DB_NAME")
	_ = viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// Redis configuration (sensitive and environment-specific)
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// Authentication (sensitive)
	_ = viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	_ = viper.BindEnv("auth.jwt_expiration", "JWT_EXPIRATION")

	// Logging (environment-specific overrides only)
	// Only bind environment variables for values that might need runtime override
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")

	// Note: Most logging, Swagger, Metrics, CORS, and Rate Limiting configurations
	// are now statically defined in config.yaml and don't need environment variable bindings
	// Only critical runtime overrides are bound above
}
