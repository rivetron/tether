package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Kosha-Nirman/tether/src/pkg/cache"
	"github.com/Kosha-Nirman/tether/src/pkg/database"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.MongoDB
	cc *cache.Redis
}

type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Version   string                   `json:"version"`
	Services  map[string]ServiceHealth `json:"services"`
}

func NewHealthHandler(db *database.MongoDB, cc *cache.Redis) *HealthHandler {
	return &HealthHandler{
		db,
		cc,
	}
}

// Health returns the overall health status
// @Summary Health check
// @Description Get the health status of the service and its dependencies
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]ServiceHealth)
	overallStatus := "healthy"

	// ? Check MongoDB
	mongoStart := time.Now()
	if err := h.db.Health(ctx); err != nil {
		services["mongodb"] = ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		services["mongodb"] = ServiceHealth{
			Status:  "healthy",
			Latency: time.Since(mongoStart).String(),
		}
	}

	// ? Check Redis
	redisStart := time.Now()
	if err := h.db.Health(ctx); err != nil {
		services["redis"] = ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		services["redis"] = ServiceHealth{
			Status:  "healthy",
			Latency: time.Since(redisStart).String(),
		}
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  services,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
