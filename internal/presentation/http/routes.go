package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/infrastructure/logger"
	"go-clean-template/internal/presentation/http/handlers"
	"go-clean-template/internal/presentation/http/middlewares"
	"go-clean-template/internal/presentation/swagger"
)

func SetupRoutes(cfg *config.Config, log logger.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middlewares.Recoverer(log))
	r.Use(middlewares.RequestLogger(log))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(middlewares.CORS(cfg.CORS))
	r.Use(middlewares.RateLimit(cfg.RateLimit))

	healthHandler := handlers.NewHealthHandler(log)

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health and monitoring endpoints
		r.Get("/health", healthHandler.Health)
		r.Get("/heartbeat", healthHandler.Heartbeat)
		r.Get("/system", healthHandler.SystemInfo)
		r.Get("/ready", healthHandler.Readiness)
		r.Get("/live", healthHandler.Liveness)
	})

	// Legacy health endpoint for backward compatibility
	r.Get("/health", healthHandler.Health)

	if cfg.Swagger.Enabled {
		log.Info("Setting up Swagger documentation",
			logger.String("route", cfg.Swagger.Route),
			logger.String("title", cfg.Swagger.Title),
		)
		swagger.SetupSwagger(r, &cfg.Swagger)
	} else {
		log.Info("Swagger documentation disabled")
	}

	return r
}
