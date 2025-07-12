package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-clean-template/internal/infrastructure/config"
	"go-clean-template/internal/shared/response"
)

// RateLimiter represents an efficient sliding window rate limiter
type RateLimiter struct {
	requests  []time.Time
	maxTokens int
	window    time.Duration
	mu        sync.RWMutex
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:  make([]time.Time, 0, maxRequests),
		maxTokens: maxRequests,
		window:    window,
	}
}

func (rl *RateLimiter) Allow() (bool, int, time.Time) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Remove requests outside the current window
	cutoff := now.Add(-rl.window)
	validIndex := 0
	for i, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validIndex = i
			break
		}
		validIndex = i + 1
	}
	rl.requests = rl.requests[validIndex:]

	remaining := rl.maxTokens - len(rl.requests)
	resetTime := now.Add(rl.window)

	// Check if request is allowed
	if remaining > 0 {
		rl.requests = append(rl.requests, now)
		return true, remaining - 1, resetTime
	}

	// Find the oldest request to determine when limit resets
	if len(rl.requests) > 0 {
		resetTime = rl.requests[0].Add(rl.window)
	}

	return false, 0, resetTime
}

// ClientLimiterStore manages rate limiters for different clients
type ClientLimiterStore struct {
	limiters    map[string]*RateLimiter
	lastCleanup time.Time
	mu          sync.RWMutex
	maxRequests int
	window      time.Duration
}

func NewClientLimiterStore(maxRequests int, window time.Duration) *ClientLimiterStore {
	return &ClientLimiterStore{
		limiters:    make(map[string]*RateLimiter),
		lastCleanup: time.Now(),
		maxRequests: maxRequests,
		window:      window,
	}
}

func (cls *ClientLimiterStore) GetLimiter(clientID string) *RateLimiter {
	// Periodic cleanup to prevent memory leaks
	if time.Since(cls.lastCleanup) > 5*time.Minute {
		cls.cleanupInactiveLimiters()
	}

	cls.mu.RLock()
	limiter, exists := cls.limiters[clientID]
	cls.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter with double-checked locking
	cls.mu.Lock()
	defer cls.mu.Unlock()

	if _, exists := cls.limiters[clientID]; exists {
		return limiter
	}

	limiter = NewRateLimiter(cls.maxRequests, cls.window)
	cls.limiters[clientID] = limiter
	return limiter
}

func (cls *ClientLimiterStore) cleanupInactiveLimiters() {
	now := time.Now()
	cls.mu.Lock()
	defer cls.mu.Unlock()

	// Remove limiters with no recent activity
	for clientID, limiter := range cls.limiters {
		limiter.mu.RLock()
		isInactive := len(limiter.requests) == 0 ||
			(len(limiter.requests) > 0 && now.Sub(limiter.requests[len(limiter.requests)-1]) > cls.window*2)
		limiter.mu.RUnlock()

		if isInactive {
			delete(cls.limiters, clientID)
		}
	}

	cls.lastCleanup = now
}

func RateLimit(rateLimitConfig config.RateLimitConfig) func(next http.Handler) http.Handler {
	// Create a store for client limiters
	store := NewClientLimiterStore(rateLimitConfig.RequestsPerMinute, time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if disabled
			if !rateLimitConfig.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get client identifier (IP address)
			clientIP := getClientIP(r)
			if clientIP == "" {
				// If we can't identify the client, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limiter for this client
			limiter := store.GetLimiter(clientIP)

			// Check if request is allowed
			allowed, remaining, resetTime := limiter.Allow()

			// Set rate limit headers (industry standard)
			setRateLimitHeaders(w, rateLimitConfig.RequestsPerMinute, remaining, resetTime)

			if !allowed {
				// Add Retry-After header
				retryAfter := int(time.Until(resetTime).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setRateLimitHeaders(w http.ResponseWriter, limit, remaining int, resetTime time.Time) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
	w.Header().Set("X-RateLimit-Window", "60") // 60 seconds window
}

func getClientIP(r *http.Request) string {
	// List of headers to check in order of preference
	headers := []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP"}

	for _, header := range headers {
		if ip := extractIPFromHeader(r.Header.Get(header)); ip != "" {
			return ip
		}
	}

	// Fall back to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if net.ParseIP(host) != nil {
			return host
		}
	}

	return r.RemoteAddr
}

func extractIPFromHeader(headerValue string) string {
	if headerValue == "" {
		return ""
	}

	// For X-Forwarded-For, take the first IP (original client)
	if idx := strings.Index(headerValue, ","); idx != -1 {
		headerValue = headerValue[:idx]
	}

	ip := strings.TrimSpace(headerValue)
	if net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}
