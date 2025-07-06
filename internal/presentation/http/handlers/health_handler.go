package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Service   string            `json:"service"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// HeartbeatResponse represents the heartbeat response
type HeartbeatResponse struct {
	Beat      string    `json:"beat"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// SystemInfoResponse represents system information
type SystemInfoResponse struct {
	GoVersion   string `json:"go_version"`
	Goroutines  int    `json:"goroutines"`
	MemoryAlloc string `json:"memory_alloc_mb"`
	MemoryTotal string `json:"memory_total_mb"`
	CPUCount    int    `json:"cpu_count"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	StartTime time.Time
	Version   string
	Service   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		StartTime: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		Service:   "go-clean-template",
	}
}

// Health provides detailed health information
//
//	@Summary		Get detailed health information
//	@Description	Returns detailed health status of the API and its dependencies
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uptime := time.Since(h.StartTime)

	// Perform health checks
	checks := make(map[string]string)
	checks["database"] = "ok" // TODO: Add actual database health check
	checks["redis"] = "ok"    // TODO: Add actual Redis health check
	checks["memory"] = "ok"

	response := HealthResponse{
		Status:    "healthy",
		Service:   h.Service,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   h.Version,
		Checks:    checks,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Heartbeat provides a simple heartbeat response
//
//	@Summary		Get heartbeat status
//	@Description	Returns a simple heartbeat response to check if the service is running
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HeartbeatResponse
//	@Router			/heartbeat [get]
func (h *HealthHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HeartbeatResponse{
		Beat:      "alive",
		Timestamp: time.Now(),
		Service:   h.Service,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// SystemInfo provides system information
//
//	@Summary		Get system information
//	@Description	Returns information about the system running the API
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	SystemInfoResponse
//	@Router			/system [get]
func (h *HealthHandler) SystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := SystemInfoResponse{
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: fmt.Sprintf("%.2f", float64(m.Alloc)/1024/1024),
		MemoryTotal: fmt.Sprintf("%.2f", float64(m.TotalAlloc)/1024/1024),
		CPUCount:    runtime.NumCPU(),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Readiness checks if the service is ready to serve traffic
//
//	@Summary		Check service readiness
//	@Description	Checks if the service is ready to serve traffic
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		503	{object}	map[string]interface{}
//	@Router			/ready [get]
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: Add actual readiness checks (database connectivity, etc.)
	ready := true

	if ready {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now(),
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "not ready",
			"timestamp": time.Now(),
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// Liveness checks if the service is alive
//
//	@Summary		Check service liveness
//	@Description	Checks if the service is alive
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/live [get]
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.StartTime).String(),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
