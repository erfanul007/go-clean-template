package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go-clean-template/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Field represents a structured logging field (abstraction over zap.Field)
type Field = zap.Field

// Logger interface defines the logging contract for the application
// Uses Field abstraction for better separation of concerns
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
}

// Convenience field constructors that wrap zap fields for backward compatibility
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Duration = zap.Duration
	Time     = zap.Time
	Error    = zap.Error
	Any      = zap.Any
	Stack    = zap.Stack
)

// zapLogger implements the Logger interface using Zap
type zapLogger struct {
	*zap.Logger
}

// LoggerConfig holds configuration for creating a logger
type LoggerConfig struct {
	Level            string
	Format           string
	EnableCaller     bool
	EnableStacktrace bool
	FileConfig       *FileConfig
}

// FileConfig holds file logging configuration
type FileConfig struct {
	Enabled    bool
	Directory  string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // Days
	Compress   bool
}

// LoggerError represents logger-specific errors
type LoggerError struct {
	Op  string
	Err error
}

func (e *LoggerError) Error() string {
	return fmt.Sprintf("logger %s: %v", e.Op, e.Err)
}

// Validate validates the logger configuration
func (c LoggerConfig) Validate() error {
	if c.Level == "" {
		return &LoggerError{Op: "validate", Err: fmt.Errorf("log level cannot be empty")}
	}
	if c.FileConfig != nil && c.FileConfig.Enabled && c.FileConfig.Directory == "" {
		return &LoggerError{Op: "validate", Err: fmt.Errorf("file logging directory cannot be empty when enabled")}
	}
	return nil
}

// New creates a new logger instance with the given configuration
func New(config LoggerConfig) (Logger, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(&config)
	ensureDockerCompatibility(&config)

	// Parse log level with fallback
	parsedLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		parsedLevel = zapcore.InfoLevel
	}

	// Create cores
	cores := []zapcore.Core{createConsoleCore(config.Format, parsedLevel)}
	if config.FileConfig != nil && config.FileConfig.Enabled {
		fileCores, err := createFileCores(config.FileConfig, parsedLevel)
		if err != nil {
			return nil, &LoggerError{Op: "create_file_cores", Err: err}
		}
		cores = append(cores, fileCores...)
	}

	// Build logger
	core := zapcore.NewTee(cores...)
	options := buildLoggerOptions(config)
	zapLog := zap.New(core, options...)

	return &zapLogger{Logger: zapLog}, nil
}

// Must creates a logger and panics on error
func Must(config LoggerConfig) Logger {
	logger, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// NewSimple creates a simple logger for basic use cases
func NewSimple(level, format string) Logger {
	return Must(LoggerConfig{
		Level:            level,
		Format:           format,
		EnableCaller:     true,
		EnableStacktrace: false,
	})
}

// NewWithConfig creates a logger from a LoggingConfig
func NewWithConfig(cfg config.LoggingConfig) (Logger, error) {
	loggerConfig := LoggerConfig{
		Level:            cfg.Level,
		Format:           cfg.Format,
		EnableCaller:     cfg.EnableCaller,
		EnableStacktrace: cfg.EnableStacktrace,
	}

	// Add file configuration if enabled
	if cfg.File.Enabled {
		loggerConfig.FileConfig = &FileConfig{
			Enabled:    cfg.File.Enabled,
			Directory:  cfg.File.Directory,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   cfg.File.Compress,
		}
	}

	return New(loggerConfig)
}

// MustWithConfig creates a logger from config and panics on error
func MustWithConfig(cfg config.LoggingConfig) Logger {
	logger, err := NewWithConfig(cfg)
	if err != nil {
		startupLogger := NewSimple(cfg.StartupLevel, cfg.StartupFormat)
		startupLogger.Fatal("Failed to initialize logger", Error(err))
	}
	return logger
}

// applyEnvironmentOverrides applies environment variable overrides to config
func applyEnvironmentOverrides(config *LoggerConfig) {
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		config.Level = envLevel
	}
	if envFormat := os.Getenv("LOG_FORMAT"); envFormat != "" {
		config.Format = envFormat
	}
}

// ensureDockerCompatibility ensures file logging uses absolute paths for Docker
func ensureDockerCompatibility(config *LoggerConfig) {
	if config.FileConfig != nil && config.FileConfig.Enabled && !filepath.IsAbs(config.FileConfig.Directory) {
		if wd, err := os.Getwd(); err == nil {
			config.FileConfig.Directory = filepath.Join(wd, config.FileConfig.Directory)
		}
	}
}

// buildLoggerOptions creates zap options based on config
func buildLoggerOptions(config LoggerConfig) []zap.Option {
	var options []zap.Option
	if config.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if config.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	return options
}

// createConsoleCore creates a console output core
func createConsoleCore(format string, level zapcore.Level) zapcore.Core {
	encoderConfig := getEncoderConfig(format)
	encoder := createEncoder(format, encoderConfig)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

// createEncoder creates an encoder based on format
func createEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	switch format {
	case "console":
		return zapcore.NewConsoleEncoder(config)
	default:
		return zapcore.NewJSONEncoder(config)
	}
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.Logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{Logger: l.Logger.With(fields...)}
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.Logger.Sync()
}

// getEncoderConfig returns encoder configuration based on format
func getEncoderConfig(format string) zapcore.EncoderConfig {
	var config zapcore.EncoderConfig

	switch format {
	case "console":
		config = zap.NewDevelopmentEncoderConfig()
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "json":
		config = zap.NewProductionEncoderConfig()
	default:
		config = zap.NewProductionEncoderConfig()
	}

	return config
}

// createFileCores creates file-based logging cores
func createFileCores(fileConfig *FileConfig, level zapcore.Level) ([]zapcore.Core, error) {
	// Ensure log directory exists
	if err := ensureLogDirectory(fileConfig); err != nil {
		return nil, err
	}

	// JSON encoder for file output
	fileEncoder := zapcore.NewJSONEncoder(getEncoderConfig("json"))

	// Single file for all logs
	allWriter := createLumberjackWriter(fileConfig, "app.log")
	allCore := zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(allWriter),
		level, // Use the provided level
	)
	return []zapcore.Core{allCore}, nil
}

// ensureLogDirectory ensures the log directory exists with proper permissions
func ensureLogDirectory(fileConfig *FileConfig) error {
	logDir := fileConfig.Directory

	// Create directory with broader permissions for Docker containers
	if err := os.MkdirAll(logDir, 0766); err != nil {
		// Try fallback directory for Docker environments
		fallbackDir := "/tmp/logs"
		if err := os.MkdirAll(fallbackDir, 0766); err != nil {
			return fmt.Errorf("failed to create log directory %s and fallback %s: %w", logDir, fallbackDir, err)
		}
		fileConfig.Directory = fallbackDir
	}
	return nil
}

// createLumberjackWriter creates a lumberjack writer for log rotation
func createLumberjackWriter(config *FileConfig, filename string) io.Writer {
	logPath := filepath.Join(config.Directory, filename)

	// Test file creation for Docker compatibility
	if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		_ = file.Close() // Ignore close error for test file
	} else {
		// Log warning but continue - lumberjack will handle the error
		fmt.Fprintf(os.Stderr, "Warning: Cannot create log file %s: %v. Logs will only go to console.\n", logPath, err)
	}

	return &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}
}

// ContextKey is the type for context keys
type ContextKey string

const LoggerContextKey ContextKey = "logger"

// FromContext extracts logger from context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(Logger); ok {
		return logger
	}
	// Return a default logger if none found in context
	return NewSimple("info", "json")
}

// WithContext adds logger to context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}
