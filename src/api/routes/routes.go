//go:build !test

package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivetron/tether/src/api/handlers"
	"github.com/rivetron/tether/src/internal/middleware"
	"github.com/rivetron/tether/src/pkg/config"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, config *config.Config, healthHandler *handlers.HealthHandler, linkHandler *handlers.LinkHandler) {
	// Health check routes
	r.GET("/health", healthHandler.Health)

	// ? Choose redirect status based on environment
	redirectStatus := http.StatusFound // * 302 for dev
	if gin.Mode() == gin.ReleaseMode {
		redirectStatus = http.StatusMovedPermanently // * 301 for prod
	}

	// ? Handle /docs (no trailing slash)
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(redirectStatus, "/docs/index.html")
	})

	// ? Handle /docs/ and all other /docs/* paths
	r.GET("/docs/*any", func(c *gin.Context) {
		any := c.Param("any")

		// * If it's just "/" (i.e., /docs/), redirect to index.html
		if any == "/" {
			c.Redirect(redirectStatus, "/docs/index.html")
			return
		}

		// * Otherwise, let swagger handle it
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	// * CORS
	corsConfig := &middleware.CORSConfig{
		AllowedOrigins: config.Security.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "X-Request-ID",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}

	// * Security
	securityConfig := &middleware.SecurityConfig{
		TrustedProxies:           config.Security.TrustedProxies,
		RequiredHeaders:          []string{},
		BlockedUserAgents:        []string{},
		EnableHTSTS:              true,
		EnableXSSProtection:      true,
		EnableContentTypeNoSniff: true,
		EnableFrameOptions:       true,
	}

	// ? Apply Middleware
	r.Use(middleware.CORSMiddleware(corsConfig))
	r.Use(middleware.SecurityMiddleware(securityConfig))

	// * API Routes with rate limiting
	api := r.Group("/api")
	api.Use(middleware.RateLimitingMiddleware())
	{
		// * Links management routes
		links := api.Group("/links")
		links.GET("", linkHandler.ListLinks)
		links.GET("/:shortCode", linkHandler.GetLink)
		links.PUT("/:shortCode", linkHandler.UpdateLink)
		links.DELETE("/:shortCode", linkHandler.DeleteLink)
		{
			links.POST("", linkHandler.CreateLink)
		}
	}

	// Redirect routes
	redirect := r.Group("/")
	{
		redirect.GET("/:shortCode", linkHandler.RedirectLink)
	}

}
