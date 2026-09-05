package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivetron/tether/src/internal/models"
	"github.com/rivetron/tether/src/internal/service"
	"github.com/rivetron/tether/src/pkg/utils"
)

type LinkHandler struct {
	linkService *service.LinkService
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type PaginationResponse struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type LinkListResponse struct {
	Links      []*models.ShortLink `json:"links"`
	Pagination PaginationResponse  `json:"pagination"`
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
	}
}

// CreateLink creates a new short link
// @Summary Create a new short link
// @Description Create a new short link with optional custom code and TTL
// @Tags Links
// @Accept json
// @Produce json
// @Param request body models.CreateLinkRequest true "Link creation request"
// @Success 201 {object} models.ShortLink
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req models.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// * Get creator IP
	creatorIP := utils.GetClientIP(c.Request)

	// * Create the Link
	response, err := h.linkService.CreateLink(c.Request.Context(), &req, creatorIP)
	if err != nil {
		status := http.StatusInternalServerError

		if err.Error() == "invalid URL format" || err.Error() == "invalid custom code format" {
			status = http.StatusBadRequest
		}

		if err.Error() == "custom code already in use" || err.Error() == "domain is not allowed" {
			status = http.StatusConflict
		}

		c.JSON(status, ErrorResponse{
			Error:   "failed to create a link",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetLink retrieves link information
// @Summary Get link information
// @Description Get detailed information about a short link
// @Tags Links
// @Produce json
// @Param shortCode path string true "Short code of the link"
// @Success 200 {object} models.ShortLink
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/links/{shortCode} [get]
func (h *LinkHandler) GetLink(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
		})
		return
	}

	link, err := h.linkService.GetLink(c.Request.Context(), shortCode)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "link not found" || err.Error() == "link has expired" {
			status = http.StatusNotFound
		}

		c.JSON(status, ErrorResponse{
			Error:   "Failed to get link",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, link)
}

// UpdateLink updates fields of an existing short link
// @Summary Update a short link
// @Description Update certain fields of a short link such as TTL or expiry
// @Tags Links
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code of the link"
// @Param request body models.UpdateLinkRequest true "Fields to update"
// @Success 200 {object} models.ShortLink
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/links/{shortCode} [put]
func (h *LinkHandler) UpdateLink(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var req models.UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	updatedLink, err := h.linkService.UpdateLink(
		c.Request.Context(),
		shortCode,
		func(link *models.ShortLink) error {

			if req.OriginalURL != nil {
				link.OriginalURL = *req.OriginalURL
			}

			if req.CustomCode != nil {
				link.CustomCode = req.CustomCode
			}

			if req.ExpiresAt != nil {
				link.ExpiresAt = req.ExpiresAt
			}

			if req.IsActive != nil {
				link.IsActive = *req.IsActive
			}

			if req.Metadata != nil {
				link.Metadata = *req.Metadata
			}

			if req.Custom != nil {
				link.Metadata.Custom = req.Custom
			}

			return nil
		},
	)

	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "link not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to update link",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedLink)
}

// DeleteLink deletes a short link
// @Summary Delete a short link
// @Description Soft delete a short link (deactivate)
// @Tags Links
// @Param shortCode path string true "Short code of the link"
// @Success 204 "Link deleted successfully"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/links/{shortCode} [delete]
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
		})
		return
	}

	err := h.linkService.DeleteLink(c.Request.Context(), shortCode)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "link not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, ErrorResponse{
			Error:   "Failed to delete link",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListLinks lists links with pagination
// @Summary List links
// @Description Get a paginated list of links
// @Tags Links
// @Produce json
// @Param limit query int false "Number of links to return" default(10)
// @Param offset query int false "Number of links to skip" default(0)
// @Success 200 {object} LinkListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/links [get]
func (h *LinkHandler) ListLinks(c *gin.Context) {
	// * Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit parameter",
			Message: "Limit must be between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset parameter",
			Message: "Offset must be non-negative",
		})
		return
	}

	// * Build filters
	filters := make(map[string]any)
	if creatorIP := c.Query("creator_ip"); creatorIP != "" {
		filters["creator_ip"] = creatorIP
	}

	links, total, err := h.linkService.ListLinks(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list links",
			Message: err.Error(),
		})
		return
	}

	response := LinkListResponse{
		Links: links,
		Pagination: PaginationResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	c.JSON(http.StatusOK, response)
}

// RedirectLink handles link redirection
// @Summary Redirect to original URL
// @Description Redirect to the original URL and record analytics
// @Tags Links
// @Param shortCode path string true "Short code of the link"
// @Success 302 "Redirect to original URL"
// @Failure 404 {object} ErrorResponse
// @Failure 410 {object} ErrorResponse "Link expired"
// @Failure 500 {object} ErrorResponse
// @Router /{shortCode} [get]
func (h *LinkHandler) RedirectLink(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
		})
		return
	}

	// * Get the link
	link, err := h.linkService.GetLink(c.Request.Context(), shortCode)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Link not found",
				Message: "The requested short link does not exist",
			})
		} else if err.Error() == "link has expired" {
			c.JSON(http.StatusGone, ErrorResponse{
				Error:   "Link expired",
				Message: "This short link has expired",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Internal error",
				Message: "Failed to process request",
			})
		}
		return
	}

	// * Increment click count (async)
	go func() {
		ctx := context.Background()

		if _, err := h.linkService.IncrementClickCount(ctx, shortCode); err != nil {
			log.Printf("Failed to increment click count: %v", err)
		}
	}()

	// * Redirect to original URL
	c.Redirect(http.StatusFound, link.OriginalURL)
}
