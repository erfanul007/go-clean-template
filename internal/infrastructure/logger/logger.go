package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go-clean-template/internal/infrastructure/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Field = zap.Field

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

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

type zapLogger struct {
	*zap.Logger
}

type LoggerConfig struct {
	Level            string
	Format           string
	EnableCaller     bool
	EnableStacktrace bool
	FileConfig       *FileConfig
}

type FileConfig struct {
	Enabled       bool
	Directory     string
	MaxSize       int // MB
	MaxBackups    int
	MaxAge        int // Days
	Compress      bool
	SeparateFiles bool
}

// New creates a new logger with the given configuration
func New(cfg LoggerConfig) (Logger, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Build cores
	cores := []zapcore.Core{
		buildConsoleCore(cfg.Format, level),
	}

	if cfg.FileConfig != nil && cfg.FileConfig.Enabled {
		if err := ensureLogDirectory(cfg.FileConfig.Directory); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		fileCores := buildFileCores(cfg.FileConfig, level)
		cores = append(cores, fileCores...)
	}

	core := zapcore.NewTee(cores...)
	options := buildOptions(cfg)
	zapLog := zap.New(core, options...)

	return &zapLogger{Logger: zapLog}, nil
}

// Must creates a logger and panics on error
func Must(cfg LoggerConfig) Logger {
	logger, err := New(cfg)
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

// NewWithConfig creates a logger from application config
func NewWithConfig(cfg config.LoggingConfig) (Logger, error) {
	loggerConfig := LoggerConfig{
		Level:            cfg.Level,
		Format:           cfg.Format,
		EnableCaller:     cfg.EnableCaller,
		EnableStacktrace: cfg.EnableStacktrace,
	}

	if cfg.File.Enabled {
		loggerConfig.FileConfig = &FileConfig{
			Enabled:       cfg.File.Enabled,
			Directory:     cfg.File.Directory,
			MaxSize:       cfg.File.MaxSize,
			MaxBackups:    cfg.File.MaxBackups,
			MaxAge:        cfg.File.MaxAge,
			Compress:      cfg.File.Compress,
			SeparateFiles: cfg.File.SeparateFiles,
		}
	}

	return New(loggerConfig)
}

// MustWithConfig creates a logger from application config and panics on error
func MustWithConfig(cfg config.LoggingConfig) Logger {
	logger, err := NewWithConfig(cfg)
	if err != nil {
		startupLogger := NewSimple(cfg.StartupLevel, cfg.StartupFormat)
		startupLogger.Fatal("Failed to initialize logger", Error(err))
	}
	return logger
}

// buildConsoleCore creates a console output core
func buildConsoleCore(format string, level zapcore.Level) zapcore.Core {
	var encoder zapcore.Encoder
	if format == "console" {
		config := zap.NewDevelopmentEncoderConfig()
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

// buildFileCores creates file output cores based on configuration
func buildFileCores(cfg *FileConfig, level zapcore.Level) []zapcore.Core {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	if !cfg.SeparateFiles {
		// Single file for all logs
		writer := &lumberjack.Logger{
			Filename:   filepath.Join(cfg.Directory, "app.log"),
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		return []zapcore.Core{
			zapcore.NewCore(encoder, zapcore.AddSync(writer), level),
		}
	}

	// Separate files for different log levels
	var cores []zapcore.Core

	// Define level-specific configurations
	levelConfigs := []struct {
		filename     string
		levelEnabler zapcore.LevelEnabler
		targetLevel  zapcore.Level
	}{
		{"error.log", levelEnabler(zapcore.ErrorLevel), zapcore.ErrorLevel},
		{"warning.log", levelEnabler(zapcore.WarnLevel), zapcore.WarnLevel},
		{"info.log", levelEnabler(zapcore.InfoLevel), zapcore.InfoLevel},
		{"debug.log", levelEnabler(zapcore.DebugLevel), zapcore.DebugLevel},
	}

	for _, config := range levelConfigs {
		if config.targetLevel >= level {
			writer := &lumberjack.Logger{
				Filename:   filepath.Join(cfg.Directory, config.filename),
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
			cores = append(cores, zapcore.NewCore(
				encoder,
				zapcore.AddSync(writer),
				config.levelEnabler,
			))
		}
	}

	return cores
}

// levelEnabler creates a level enabler for a specific level only
func levelEnabler(targetLevel zapcore.Level) zapcore.LevelEnabler {
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == targetLevel
	})
}

// buildOptions creates zap options based on configuration
func buildOptions(cfg LoggerConfig) []zap.Option {
	var options []zap.Option
	if cfg.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	return options
}

// ensureLogDirectory creates the log directory if it doesn't exist
func ensureLogDirectory(dir string) error {
	if dir == "" {
		dir = getDefaultLogDirectory()
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		// Fallback to ./logs if the specified directory fails
		if fallbackErr := os.MkdirAll("./logs", 0755); fallbackErr != nil {
			return fmt.Errorf("failed to create log directory: %w", fallbackErr)
		}
	}
	return nil
}

// getDefaultLogDirectory returns the appropriate default log directory
func getDefaultLogDirectory() string {
	if os.Getenv("DOCKER_CONTAINER") != "" || os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "/app/logs"
	}
	return "./logs"
}

// validate checks if the configuration is valid
func (c LoggerConfig) validate() error {
	if c.Level == "" {
		return fmt.Errorf("log level cannot be empty")
	}
	if c.FileConfig != nil && c.FileConfig.Enabled && c.FileConfig.Directory == "" {
		return fmt.Errorf("file logging directory cannot be empty when enabled")
	}
	return nil
}

// Logger interface implementations
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.Logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.Logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{Logger: l.Logger.With(fields...)}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	return &zapLogger{Logger: l.Logger.With(zap.String("trace_id", getTraceID(ctx)))}
}

func (l *zapLogger) Sync() error {
	return l.Logger.Sync()
}

// Context utilities
type ContextKey string

const LoggerContextKey ContextKey = "logger"

// FromContext retrieves a logger from context or returns a default one
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(Logger); ok {
		return logger
	}
	return NewSimple("info", "json")
}

// WithContext adds a logger to context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// getTraceID extracts trace ID from context
func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return "unknown"
}
