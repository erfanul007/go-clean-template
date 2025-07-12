package main

import (
	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/infrastructure/logger"
	"go-clean-template/internal/presentation/http"
	"go-clean-template/internal/presentation/swagger"
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
	cfg, err := config.Load()
	if err != nil {
		log := logger.NewSimple("error", "console")
		log.Fatal("Failed to load configuration", logger.Error(err))
	}

	log := logger.MustWithConfig(cfg.Logging)
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Application starting",
		logger.String("environment", cfg.Server.Environment),
		logger.String("version", "1.0.0"),
		logger.String("port", cfg.Server.Port),
	)

	swagger.Initialize(cfg.Swagger)

	server := http.NewServer(cfg, log)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start", logger.Error(err))
	}
}
