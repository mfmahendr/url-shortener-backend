package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestAnalytics(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	require.NotNil(t, tcEnv, "tcEnv should be initialized")

	// Middleware & controller
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)
	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(5, 5*time.Second)

	// services and controller
	urlSvc := url_service.New(fsService, fsService, nil)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)
	controller := controllers.New(urlSvc, trackingSvc, fsService, rateLimiter)
	controller.Router.GET("/u/analytics/:short_id",
		rateLimiter.Apply(authMiddleware.RequireAuth(controller.Analytics)),
	)

	// Buat data user dan shortlink
	ownerUID, ownerToken := createTestUserAndToken(t, "analytics-owner@example.com")
	_, anotherToken := createTestUserAndToken(t, "not-owner@example.com")

	shortID := "analyticsTest123"
	err := fsService.SetShortlink(ctx, shortID, models.Shortlink{
		ShortID:   shortID,
		URL:       "https://example.com/analytics",
		CreatedBy: ownerUID,
		IsPrivate: true,
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// Simulate click logs
	_ = fsService.AddClickLog(ctx, &models.ClickLog{
		ShortID:   shortID,
		Timestamp: time.Now(),
		IP:        "192.0.2.1",
		UserAgent: "Mozilla/5.0 (compatible; TestAnalytics)",
	})

	t.Run("success fetch analytics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/analytics/"+shortID, nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"short_id":"`+shortID+`"`)
		assert.Contains(t, rec.Body.String(), `"ip":"192.0.2.1"`)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/analytics/"+shortID, nil)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing or invalid Authorization header")
	})

	t.Run("forbidden access by non-owner", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/analytics/"+shortID, nil)
		req.Header.Set("Authorization", "Bearer "+anotherToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Forbidden")
	})

	t.Run("invalid short_id format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/analytics/###", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("not found short_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/analytics/notfound123", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrNotFound.Error())
	})
}
