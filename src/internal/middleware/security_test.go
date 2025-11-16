package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware_DefaultHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityMiddleware(nil)) // use default
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "1; mode=block", resp.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", resp.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header().Get("X-Frame-Options"))

	assert.Equal(t, "none", resp.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header().Get("Referrer-Policy"))

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSecurityMiddleware_RequiredHeadersMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &SecurityConfig{
		RequiredHeaders: []string{"X-Test-Header"},
	}

	router := gin.New()
	router.Use(SecurityMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "Required header missing: X-Test-Header")
}

func TestSecurityMiddleware_RequiredHeadersPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &SecurityConfig{
		RequiredHeaders: []string{"X-Test-Header"},
	}

	router := gin.New()
	router.Use(SecurityMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-Header", "123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSecurityMiddleware_BlockedUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &SecurityConfig{
		BlockedUserAgents: []string{"badbot"},
	}

	router := gin.New()
	router.Use(SecurityMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "SuperBadBot v1.0")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Contains(t, strings.ToLower(resp.Body.String()), "access denied")
}

func TestSecurityMiddleware_AllowedUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &SecurityConfig{
		BlockedUserAgents: []string{"badbot"},
	}

	router := gin.New()
	router.Use(SecurityMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
