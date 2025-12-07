package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SecurityConfig struct {
	TrustedProxies           []string
	RequiredHeaders          []string
	BlockedUserAgents        []string
	EnableHTSTS              bool
	EnableXSSProtection      bool
	EnableContentTypeNoSniff bool
	EnableFrameOptions       bool
}

func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc {
	if config == nil {
		config = defaultSecurityConfig()
	}

	return func(ctx *gin.Context) {
		// ? Set security headers
		if config.EnableHTSTS {
			ctx.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		if config.EnableXSSProtection {
			ctx.Header("X-XSS-Protection", "1; mode=block")
		}

		if config.EnableContentTypeNoSniff {
			ctx.Header("X-Content-Type-Options", "nosniff")
		}

		if config.EnableFrameOptions {
			ctx.Header("X-Frame-Options", "DENY")
		}

		// ? Set additional security headers
		ctx.Header("X-Permitted-Cross-Domain-Policies", "none")
		ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Check required headers
		for _, header := range config.RequiredHeaders {
			if ctx.GetHeader(header) == "" {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": "Required header missing: " + header,
				})
				ctx.Abort()
				return
			}
		}

		// Check blocked user agents
		userAgent := ctx.GetHeader("User-Agent")
		for _, blocked := range config.BlockedUserAgents {
			if strings.Contains(strings.ToLower(userAgent), strings.ToLower(blocked)) {
				ctx.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
				})
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}

// TODO: Implement proper rate limiting
func RateLimitingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. Getting client IP
		// 2. Checking rate limit in Redis/memory store
		// 3. Updating counters
		// 4. Blocking if limit exceeded
		ctx.Next()
	}
}

func defaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		TrustedProxies:           []string{},
		RequiredHeaders:          []string{},
		BlockedUserAgents:        []string{},
		EnableHTSTS:              true,
		EnableXSSProtection:      true,
		EnableContentTypeNoSniff: true,
		EnableFrameOptions:       true,
	}
}
