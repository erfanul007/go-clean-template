package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/infrastructure/logger"
)

type Server struct {
	server *http.Server
	config *config.Config
	logger logger.Logger
}

func NewServer(config *config.Config, log logger.Logger) *Server {
	// Setup routes with configuration and logger
	router := SetupRoutes(config, log)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Server.WriteTimeout) * time.Second,
	}

	return &Server{
		server: server,
		config: config,
		logger: log,
	}
}

func (s *Server) Start() error {
	go func() {
		s.logger.Info("HTTP server starting",
			logger.String("port", s.config.Server.Port),
			logger.String("host", s.config.Server.Host),
			logger.String("environment", s.config.Server.Environment),
			logger.Duration("read_timeout", time.Duration(s.config.Server.ReadTimeout)*time.Second),
			logger.Duration("write_timeout", time.Duration(s.config.Server.WriteTimeout)*time.Second),
		)

		if s.config.Swagger.Enabled {
			s.logger.Info("Swagger UI available",
				logger.String("url", fmt.Sprintf("http://%s:%s/swagger/index.html", s.config.Server.Host, s.config.Server.Port)),
			)
		}

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("HTTP server failed to start", logger.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	s.logger.Info("Initiating graceful server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", logger.Error(err))
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info("Server shutdown completed successfully")
	return nil
}
