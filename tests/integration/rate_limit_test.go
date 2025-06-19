package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)

	// Middleware
	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(3, 3*time.Second) // allow 3 requests per 3 seconds

	// Dummy controller with limited endpoint
	controller := controllers.New(nil, nil, nil, rateLimiter)
	controller.Router.GET("/health", rateLimiter.Apply(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	t.Run("hit limit after threshold", func(t *testing.T) {
		for i := 1; i <= 4; i++ {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			req.RemoteAddr = "127.0.0.1:1234" // same IP
			rec := httptest.NewRecorder()

			controller.Router.ServeHTTP(rec, req)

			if i <= 3 {
				assert.Equal(t, http.StatusOK, rec.Code, "request #%d should succeed", i)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, rec.Code, "request #%d should be rate limited", i)
			}
		}
	})
}
