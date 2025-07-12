package handlers

import (
	"net/http"
	"runtime"
	"time"

	"go-clean-template/internal/infrastructure/logger"
	"go-clean-template/internal/shared/response"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	logger logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: log,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SystemInfoResponse represents system information
type SystemInfoResponse struct {
	Status       string            `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Service      string            `json:"service"`
	Version      string            `json:"version"`
	GoVersion    string            `json:"go_version"`
	NumCPU       int               `json:"num_cpu"`
	NumGoroutine int               `json:"num_goroutine"`
	Memory       map[string]uint64 `json:"memory"`
	Uptime       string            `json:"uptime"`
}

var startTime = time.Now()

// Health returns basic health status
// @Summary Get health status
// @Description Returns the basic health status of the service
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Health check endpoint called",
		logger.String("method", r.Method),
		logger.String("path", r.URL.Path),
	)

	response.Success(w, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "go-clean-template",
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
	})

	h.logger.Debug("Health check completed successfully")
}

// Heartbeat returns a simple heartbeat response
// @Summary Get heartbeat
// @Description Returns a simple heartbeat response to verify service is alive
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /heartbeat [get]
func (h *HealthHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Heartbeat endpoint called")

	response.Success(w, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"service":   "go-clean-template",
	})
}

// SystemInfo returns detailed system information
// @Summary Get system information
// @Description Returns detailed system information including memory usage, CPU count, and runtime stats
// @Tags Health
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Router /system [get]
func (h *HealthHandler) SystemInfo(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("System info endpoint called")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response.Success(w, SystemInfoResponse{
		Status:       "healthy",
		Timestamp:    time.Now(),
		Service:      "go-clean-template",
		Version:      "1.0.0",
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Memory: map[string]uint64{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      uint64(m.NumGC),
		},
		Uptime: time.Since(startTime).String(),
	})

	h.logger.Debug("System info completed successfully")
}

// Readiness checks if the service is ready to serve requests
// @Summary Get readiness status
// @Description Checks if the service is ready to serve requests by verifying dependencies
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Readiness check endpoint called")

	// In a real application, you would check dependencies here
	// e.g., database connectivity, external services, etc.
	checks := map[string]string{
		"database": "healthy", // This would be a real check
		"redis":    "healthy", // This would be a real check
		"storage":  "healthy", // This would be a real check
	}

	response.Success(w, HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Service:   "go-clean-template",
		Version:   "1.0.0",
		Checks:    checks,
	})

	h.logger.Debug("Readiness check completed successfully")
}

// Liveness checks if the service is alive
// @Summary Get liveness status
// @Description Checks if the service is alive and responding
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /live [get]
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Liveness check endpoint called")

	response.Success(w, HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
		Service:   "go-clean-template",
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
	})

	h.logger.Debug("Liveness check completed successfully")
}
