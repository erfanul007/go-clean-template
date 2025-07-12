package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

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

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Environment  string `mapstructure:"environment"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	JWTExpiration int    `mapstructure:"jwt_expiration"`
}

type LoggingConfig struct {
	Level            string            `mapstructure:"level"`
	Format           string            `mapstructure:"format"`
	StartupLevel     string            `mapstructure:"startup_level"`
	StartupFormat    string            `mapstructure:"startup_format"`
	EnableCaller     bool              `mapstructure:"enable_caller"`
	EnableStacktrace bool              `mapstructure:"enable_stacktrace"`
	File             FileLoggingConfig `mapstructure:"file"`
}

type FileLoggingConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Directory     string `mapstructure:"directory"`
	MaxSize       int    `mapstructure:"max_size"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxAge        int    `mapstructure:"max_age"`
	Compress      bool   `mapstructure:"compress"`
	SeparateFiles bool   `mapstructure:"separate_files"`
}

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

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	setDefaults()
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := loadEnvFile(); err != nil {
		log.Printf("Config warning: %v", err)
	}

	mapEnvVariables()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("auth.jwt_expiration", 3600)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}

func loadEnvFile() error {
	possibleLocations := []string{
		".env",
		filepath.Join(".", ".env"),
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	for _, location := range possibleLocations {
		if _, err := os.Stat(location); err == nil {
			err := godotenv.Load(location)
			if err != nil {
				return fmt.Errorf("error loading .env file from %s: %w", location, err)
			}
			return nil
		}
	}

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
			return nil
		}
	}

	return nil
}

func mapEnvVariables() {
	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("server.host", "HOST")
	_ = viper.BindEnv("server.environment", "ENVIRONMENT")

	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.dbname", "DB_NAME")
	_ = viper.BindEnv("database.sslmode", "DB_SSLMODE")

	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	_ = viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	_ = viper.BindEnv("auth.jwt_expiration", "JWT_EXPIRATION")

	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")
	_ = viper.BindEnv("logging.file.separate_files", "LOG_SEPARATE_FILES")
}
