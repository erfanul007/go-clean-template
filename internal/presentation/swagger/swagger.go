package swagger

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"go-clean-template/internal/infrastructure/config"

	// Import generated docs
	_ "go-clean-template/docs"
)

// SwaggerConfig is an alias for the config.SwaggerConfig to maintain backward compatibility
type SwaggerConfig = config.SwaggerConfig

// GetSwaggerConfig returns the Swagger configuration from the global config
func GetSwaggerConfig() (*SwaggerConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &cfg.Swagger, nil
}

// SetupSwagger configures Swagger documentation for the API
func SetupSwagger(r *chi.Mux, swaggerConfig *SwaggerConfig) {
	if !swaggerConfig.Enabled {
		return
	}

	// Add redirect for /swagger and /swagger/ to /swagger/index.html
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Mount Swagger UI handler
	// The wildcard route pattern is important for the Swagger UI to work correctly
	r.Get(swaggerConfig.Route, httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))
}
