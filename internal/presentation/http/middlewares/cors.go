package middlewares

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-clean-template/internal/infrastructure/config"
)

func CORS(corsConfig config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			isPreflightRequest := r.Method == http.MethodOptions

			// Check if origin is allowed (single validation)
			originAllowed := origin != "" && len(corsConfig.AllowedOrigins) > 0 && isOriginAllowed(origin, corsConfig.AllowedOrigins)

			// Set CORS headers for allowed origins or preflight requests
			if originAllowed || isPreflightRequest {
				setCORSHeaders(w, corsConfig, origin, originAllowed)
			}

			// Handle preflight requests
			if isPreflightRequest {
				if originAllowed {
					w.WriteHeader(http.StatusNoContent)
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setCORSHeaders(w http.ResponseWriter, corsConfig config.CORSConfig, origin string, originAllowed bool) {
	// Set origin header only for allowed origins
	if originAllowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if corsConfig.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	// Set other headers for allowed origins or preflight requests
	if originAllowed {
		if len(corsConfig.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
		}
		if len(corsConfig.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
		}
		if len(corsConfig.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposedHeaders, ", "))
		}
		if corsConfig.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
		}
	}
}

// isOriginAllowed securely checks if the origin is in the allowed origins list
// Implements security best practices to prevent CORS bypass attacks
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	// Security: Reject null origin to prevent sandboxed iframe attacks
	if origin == "null" {
		return false
	}

	// Security: Validate origin format to prevent malformed URLs
	parsedOrigin, err := url.Parse(origin)
	if err != nil || parsedOrigin.Scheme == "" || parsedOrigin.Host == "" {
		return false
	}

	// Security: Only allow http/https schemes
	if parsedOrigin.Scheme != "http" && parsedOrigin.Scheme != "https" {
		return false
	}

	for _, allowed := range allowedOrigins {
		// Security: Never allow wildcard (*) - this is a major security vulnerability
		if allowed == "*" {
			continue
		}

		// Exact match (most secure)
		if allowed == origin {
			return true
		}

		// Subdomain support with proper validation
		if strings.HasPrefix(allowed, "*.") {
			if isSubdomainMatch(parsedOrigin, allowed) {
				return true
			}
		}
	}
	return false
}

func isSubdomainMatch(parsedOrigin *url.URL, allowedPattern string) bool {
	domain := strings.TrimPrefix(allowedPattern, "*.")
	if domain == "" {
		return false
	}

	// Parse the allowed domain to ensure it's valid
	allowedURL, err := url.Parse("https://" + domain)
	if err != nil || allowedURL.Host == "" {
		return false
	}

	// Handle scheme-specific patterns
	if strings.HasPrefix(allowedPattern, "http://") {
		domain = strings.TrimPrefix(allowedPattern, "http://*.")
		if parsedOrigin.Scheme != "http" {
			return false
		}
	} else if strings.HasPrefix(allowedPattern, "https://") {
		domain = strings.TrimPrefix(allowedPattern, "https://*.")
		if parsedOrigin.Scheme != "https" {
			return false
		}
	}

	// Security: Check subdomain with proper boundary validation
	// This prevents attacks like "attackerexample.com" matching "*.example.com"
	return parsedOrigin.Host == domain || strings.HasSuffix(parsedOrigin.Host, "."+domain)
}
