package middlewares

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"go-clean-template/internal/infrastructure/logger"
)

// Skip patterns for reducing log noise
var (
	skipPaths = map[string]bool{
		"/health":      true,
		"/healthz":     true,
		"/ping":        true,
		"/metrics":     true,
		"/favicon.ico": true,
	}

	skipPrefixes = []string{"/static/", "/assets/"}
	skipSuffixes = []string{".css", ".js", ".ico"}

	correlationHeaders = []string{
		"X-Correlation-ID", "X-Correlation-Id",
		"X-Request-ID", "X-Request-Id",
		"X-Trace-ID", "X-Trace-Id",
	}

	ipHeaders = []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP"}
)

// RequestLogger creates a structured logging middleware following industry best practices
// Logs HTTP requests with correlation ID, performance metrics, and security-relevant information
func RequestLogger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipLogging(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			ctx := extractRequestContext(r)

			next.ServeHTTP(ww, r)

			// Build log fields with metrics
			fields := buildRequestFields(r, ctx, ww, time.Since(start))

			// Log with appropriate level based on status
			logWithLevel(log, ww.Status(), "HTTP Request", fields...)
		})
	}
}

// Recoverer creates a panic recovery middleware with comprehensive error logging
// Follows industry best practices for panic recovery and structured error reporting
func Recoverer(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					ctx := extractRequestContext(r)
					fields := buildPanicFields(r, ctx, rvr)

					log.Error("Panic recovered - Critical Error", fields...)

					if !isResponseWritten(w) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(`{"error":"Internal Server Error","message":"An unexpected error occurred"}`))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions for logging middleware

// requestContext holds common request information
type requestContext struct {
	correlationID string
	requestID     string
	clientIP      string
}

// shouldSkipLogging determines if a request should be skipped from logging
func shouldSkipLogging(path string) bool {
	if skipPaths[path] {
		return true
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	for _, suffix := range skipSuffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	return false
}

// extractRequestContext extracts common request context information
func extractRequestContext(r *http.Request) requestContext {
	return requestContext{
		correlationID: getHeaderValue(r, correlationHeaders, middleware.GetReqID(r.Context())),
		requestID:     middleware.GetReqID(r.Context()),
		clientIP:      extractClientIP(r),
	}
}

// getHeaderValue extracts the first non-empty header value from a list
func getHeaderValue(r *http.Request, headers []string, fallback string) string {
	for _, header := range headers {
		if value := r.Header.Get(header); value != "" {
			return value
		}
	}
	return fallback
}

// extractClientIP extracts the real client IP address from request
func extractClientIP(r *http.Request) string {
	// Check proxy headers in order
	for _, header := range ipHeaders {
		if value := r.Header.Get(header); value != "" {
			// For X-Forwarded-For, take the first IP
			if header == "X-Forwarded-For" {
				if ips := strings.Split(value, ","); len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return value
		}
	}

	// Fallback to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}
	return r.RemoteAddr
}

// buildRequestFields creates log fields for HTTP requests
func buildRequestFields(r *http.Request, ctx requestContext, ww middleware.WrapResponseWriter, duration time.Duration) []logger.Field {
	fields := []logger.Field{
		logger.String("method", r.Method),
		logger.String("path", r.URL.Path),
		logger.String("correlation_id", ctx.correlationID),
		logger.String("request_id", ctx.requestID),
		logger.String("client_ip", ctx.clientIP),
		logger.Int("status", ww.Status()),
		logger.Int("response_size_bytes", ww.BytesWritten()),
		logger.Duration("duration_ms", duration),
		logger.String("user_agent", r.UserAgent()),
	}

	if r.URL.RawQuery != "" {
		fields = append(fields, logger.String("query", r.URL.RawQuery))
	}

	if r.ContentLength > 0 {
		fields = append(fields, logger.Int64("request_size_bytes", r.ContentLength))
	}

	return fields
}

// buildPanicFields creates log fields for panic recovery
func buildPanicFields(r *http.Request, ctx requestContext, panicValue interface{}) []logger.Field {
	return []logger.Field{
		logger.String("method", r.Method),
		logger.String("path", r.URL.Path),
		logger.String("correlation_id", ctx.correlationID),
		logger.String("request_id", ctx.requestID),
		logger.String("client_ip", ctx.clientIP),
		logger.String("user_agent", r.UserAgent()),
		logger.Any("panic_value", panicValue),
		logger.String("stack_trace", string(debug.Stack())),
	}
}

// logWithLevel logs with appropriate level based on HTTP status code
func logWithLevel(log logger.Logger, status int, message string, fields ...logger.Field) {
	switch {
	case status >= 500:
		log.Error(message, fields...)
	case status >= 400:
		log.Warn(message, fields...)
	default:
		log.Info(message, fields...)
	}
}

// isResponseWritten checks if the response has already been written
func isResponseWritten(w http.ResponseWriter) bool {
	if rw, ok := w.(interface{ Status() int }); ok {
		return rw.Status() != 0
	}
	return false
}
