//go:build !test

package routes

import (
	"net/http"

	"github.com/Kosha-Nirman/tether/src/api/handlers"
	"github.com/Kosha-Nirman/tether/src/pkg/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, config *config.Config, healthHandler *handlers.HealthHandler) {
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
}
