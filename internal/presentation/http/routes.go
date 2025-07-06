package http

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go-clean-template/internal/presentation/http/handlers"
	"go-clean-template/internal/presentation/swagger"
)

// SetupRoutes configures all API routes
func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health and monitoring endpoints
		r.Get("/health", healthHandler.Health)
		r.Get("/heartbeat", healthHandler.Heartbeat)
		r.Get("/system", healthHandler.SystemInfo)
		r.Get("/ready", healthHandler.Readiness)
		r.Get("/live", healthHandler.Liveness)

		// Future routes will be added here:
		// - /api/v1/entities/*
		// - /api/v1/resources/*
		// - /api/v1/services/*
	})

	// Legacy health endpoint for backward compatibility
	r.Get("/health", healthHandler.Health)

	// Setup Swagger UI using the swagger package with configuration from global config
	swaggerConfig, err := swagger.GetSwaggerConfig()
	if err != nil {
		// If there's an error loading the config, log it and continue without Swagger
		// In a production application, you might want to handle this differently
		log.Printf("Failed to load Swagger configuration: %v", err)
		return r
	}
	swagger.SetupSwagger(r, swaggerConfig)

	return r
}
