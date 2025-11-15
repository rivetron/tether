package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = defaultCORSConfig()
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")

		// ? Check if origin is allowed
		if isOriginAllowed(origin, config.AllowedOrigins) {
			ctx.Header("Access-Control-Allow-Origin", origin)
		} else if contains(config.AllowedOrigins, "*") {
			ctx.Header("Access-Control-Allow-Origin", "*")
		}

		// ? Set other CORS Headers
		ctx.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		ctx.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

		if len(config.ExposedHeaders) > 0 {
			ctx.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		if config.AllowCredentials {
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			ctx.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// ? Handle preflight requests
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

func defaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},
		AllowCredentials: false,
		MaxAge:           86400, // ? 24 Hours
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}

		// * Support for wildcard subdomains like *.example.com
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, "."+domain) {
				return true
			}
		}
	}

	return false
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
