package main

import (
	"go-clean-template/docs"
	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/infrastructure/logger"
	"go-clean-template/internal/presentation/http"
)

//	@title			Go Clean Architecture API
//	@version		1.0
//	@description	A Go API template built with Clean Architecture and Domain-Driven Design (DDD) principles.

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

// @BasePath	/api/v1
// @schemes	http https
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use a basic logger for startup errors (with defaults since config failed to load)
		startupLogger := logger.NewSimple("error", "console")
		startupLogger.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Initialize logger from configuration (abstracted to logger package)
	log := logger.MustWithConfig(cfg.Logging)
	defer func() {
		// Ignore sync errors on stdout/stderr as they're expected
		_ = log.Sync()
	}()

	log.Info("Application starting",
		logger.String("environment", cfg.Server.Environment),
		logger.String("version", "1.0.0"),
		logger.String("port", cfg.Server.Port),
	)

	// Set Swagger info
	setSwaggerInfo(cfg)

	// Create and start server
	server := http.NewServer(cfg, log)
	if err := server.Start(); err != nil {
		log.Fatal("Server startup failed", logger.Error(err))
	}
}

// setSwaggerInfo updates Swagger info based on configuration
func setSwaggerInfo(cfg *config.Config) {
	// Set Swagger info from configuration
	docs.SwaggerInfo.Title = cfg.Swagger.Title
	docs.SwaggerInfo.Description = cfg.Swagger.Description
	docs.SwaggerInfo.Version = cfg.Swagger.Version
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = cfg.Swagger.BasePath
	docs.SwaggerInfo.Schemes = cfg.Swagger.Schemes
}
