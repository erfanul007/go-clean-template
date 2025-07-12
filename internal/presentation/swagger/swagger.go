package swagger

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"go-clean-template/docs"
	"go-clean-template/internal/infrastructure/config"
)

type SwaggerConfig = config.SwaggerConfig

func Initialize(cfg SwaggerConfig) {
	docs.SwaggerInfo.Title = cfg.Title
	docs.SwaggerInfo.Description = cfg.Description
	docs.SwaggerInfo.Version = cfg.Version
	docs.SwaggerInfo.Host = cfg.Host
	docs.SwaggerInfo.BasePath = cfg.BasePath
	docs.SwaggerInfo.Schemes = cfg.Schemes
}

func SetupSwagger(r *chi.Mux, swaggerConfig *SwaggerConfig) {
	if !swaggerConfig.Enabled {
		return
	}

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Mount Swagger UI handler
	r.Get(swaggerConfig.Route, httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))
}
