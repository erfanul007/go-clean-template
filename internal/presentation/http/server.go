package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-clean-template/internal/infrastructure/config"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.Config
}

// NewServer creates a new HTTP server
func NewServer(config *config.Config) *Server {
	router := SetupRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Server.WriteTimeout) * time.Second,
	}

	return &Server{
		server: server,
		config: config,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", s.config.Server.Port)
		log.Printf("Swagger UI available at http://%s:%s/swagger/index.html", s.config.Server.Host, s.config.Server.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited properly")
	return nil
}
