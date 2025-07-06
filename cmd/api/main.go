package main

import (
	"log"

	"go-clean-template/docs"
	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/presentation/http"
)

//	@title			Go Clean Architecture API
//	@version		1.0
//	@description	A Go API template built with Clean Architecture and Domain-Driven Design (DDD) principles.

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api/v1
//	@schemes	http https
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Swagger info
	setSwaggerInfo()

	// Create and start server
	server := http.NewServer(cfg)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// setSwaggerInfo updates Swagger info based on configuration
func setSwaggerInfo() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load configuration for Swagger: %v, using defaults", err)
		return
	}

	// Set Swagger info from configuration
	docs.SwaggerInfo.Title = cfg.Swagger.Title
	docs.SwaggerInfo.Description = cfg.Swagger.Description
	docs.SwaggerInfo.Version = cfg.Swagger.Version
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = cfg.Swagger.BasePath
	docs.SwaggerInfo.Schemes = cfg.Swagger.Schemes
}
