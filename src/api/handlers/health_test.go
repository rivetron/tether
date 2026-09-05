package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rivetron/tether/src/api/handlers"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	err error
}

func (m *mockDB) Health(ctx context.Context) error { return m.err }

type mockCache struct {
	err error
}

func (m *mockCache) Health(ctx context.Context) error { return m.err }

func setupRouter(h *handlers.HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", h.Health)
	return r
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		mongoErr       error
		redisErr       error
		expectedStatus int
		expectedSvc    map[string]string
	}{
		{
			name:           "all healthy",
			mongoErr:       nil,
			redisErr:       nil,
			expectedStatus: http.StatusOK,
			expectedSvc: map[string]string{
				"mongodb": "healthy",
				"redis":   "healthy",
			},
		},
		{
			name:           "mongo unhealthy",
			mongoErr:       errors.New("mongo down"),
			redisErr:       nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedSvc: map[string]string{
				"mongodb": "unhealthy",
				"redis":   "healthy",
			},
		},
		{
			name:           "redis unhealthy",
			mongoErr:       nil,
			redisErr:       errors.New("redis down"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedSvc: map[string]string{
				"mongodb": "healthy",
				"redis":   "unhealthy",
			},
		},
		{
			name:           "both unhealthy",
			mongoErr:       errors.New("mongo down"),
			redisErr:       errors.New("redis down"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedSvc: map[string]string{
				"mongodb": "unhealthy",
				"redis":   "unhealthy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{err: tt.mongoErr}
			mockCache := &mockCache{err: tt.redisErr}

			h := handlers.NewHealthHandler(mockDB, mockCache)
			router := setupRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			body := w.Body.String()

			for svc, status := range tt.expectedSvc {
				assert.Contains(t, body, svc)
				assert.Contains(t, body, status)
			}
		})
	}
}
