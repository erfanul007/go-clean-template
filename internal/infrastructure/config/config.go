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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
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
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int `mapstructure:"requests"`
	Window   int `mapstructure:"window"`
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

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "app_db")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-super-secret-jwt-key-change-this-in-production")
	viper.SetDefault("auth.jwt_expiration", 3600)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Swagger defaults
	viper.SetDefault("swagger.enabled", true)
	viper.SetDefault("swagger.route", "/swagger/*")
	viper.SetDefault("swagger.title", "Go Clean Architecture API")
	viper.SetDefault("swagger.description", "A comprehensive API template built with Go and Clean Architecture")
	viper.SetDefault("swagger.version", "1.0")
	viper.SetDefault("swagger.host", "localhost:8080")
	viper.SetDefault("swagger.base_path", "/api/v1")
	viper.SetDefault("swagger.schemes", []string{"http", "https"})

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", "9090")

	// Rate limiting defaults
	viper.SetDefault("rate_limit.requests", 100)
	viper.SetDefault("rate_limit.window", 60)
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
func mapEnvVariables() {
	// Server
	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("server.host", "HOST")
	_ = viper.BindEnv("server.environment", "ENVIRONMENT")

	// Database
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.dbname", "DB_NAME")
	_ = viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// Redis
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// Auth
	_ = viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	_ = viper.BindEnv("auth.jwt_expiration", "JWT_EXPIRATION")

	// Logging
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")

	// Swagger - typically not set via env vars but could be
	_ = viper.BindEnv("swagger.enabled", "SWAGGER_ENABLED")
	_ = viper.BindEnv("swagger.host", "SWAGGER_HOST")
	_ = viper.BindEnv("swagger.base_path", "SWAGGER_BASE_PATH")

	// CORS
	_ = viper.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
	_ = viper.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS")
	_ = viper.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS")

	// Metrics
	_ = viper.BindEnv("metrics.enabled", "METRICS_ENABLED")
	_ = viper.BindEnv("metrics.port", "METRICS_PORT")

	// Rate limiting
	_ = viper.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	_ = viper.BindEnv("rate_limit.window", "RATE_LIMIT_WINDOW")
}
