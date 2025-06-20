package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetClickCount(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	require.NotNil(t, tcEnv, "tcEnv should be initialized")

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)

	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(5, 5*time.Second)

	// services and controller
	urlSvc := url_service.New(fsService, fsService, nil)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)
	controller := controllers.New(urlSvc, trackingSvc, fsService, rateLimiter)
	controller.Router.GET("/u/click-count/:short_id",
		rateLimiter.Apply(authMiddleware.RequireAuth(controller.GetClickCount)),
	)

	// Create test user & shortlink
	uid, token, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "clickuser@example.com", nil)
	require.NoError(t, err)
	shortID := "clickcount123"

	// shortlink creation
	err = fsService.SetShortlink(ctx, shortID, models.Shortlink{
		ShortID:   shortID,
		URL:       "https://clickcount.com",
		CreatedBy: uid,
		IsPrivate: true,
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// Set initial click count in Redis
	err = tcEnv.rdClient.Set(ctx, "clicks:"+shortID, 42, 0).Err()
	require.NoError(t, err)

	t.Run("success get click count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res dto.ClickCountResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, int64(42), res.ClickCount)
		assert.Equal(t, shortID, res.ShortID)
	})

	t.Run("unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID, nil)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing or invalid Authorization header")
	})

	t.Run("forbidden access by non-owner", func(t *testing.T) {
		_, otherToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "anotheruser@example.com", nil)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID, nil)
		req.Header.Set("Authorization", "Bearer "+otherToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Forbidden")
	})

	t.Run("not found shortlink", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/notfound123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid short ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/@@@", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})
}

func TestExportAllClickCount(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	require.NotNil(t, tcEnv, "tcEnv should be initialized")

	// middleware + controller setup
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)
	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(5, 5*time.Second)

	// controller setup
	urlSvc := url_service.New(fsService, fsService, nil)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)
	controller := controllers.New(urlSvc, trackingSvc, fsService, rateLimiter)
	controller.Router.GET("/u/click-count/:short_id/export",
		rateLimiter.Apply(authMiddleware.RequireAuth(controller.ExportAllClickCount)),
	)

	// create test user and shortlink
	ownerUID, ownerToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "ownerexport@example.com", nil)
	require.NoError(t, err)
	_, otherToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "otherexport@example.com", nil)
	require.NoError(t, err)

	shortID := "exporttest123"
	err = fsService.SetShortlink(ctx, shortID, models.Shortlink{
		ShortID:   shortID,
		URL:       "https://export.com",
		CreatedBy: ownerUID,
		CreatedAt: time.Now(),
		IsPrivate: true,
	})
	require.NoError(t, err)

	// Simulate click logs
	_ = fsService.AddClickLog(ctx, &models.ClickLog{
		ShortID:   shortID,
		Timestamp: time.Now(),
		IP:        "127.0.0.1",
		UserAgent: "TestAgent/1.0",
	})

	t.Run("success export CSV", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID+"/export", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/csv", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "analytics_"+shortID+".csv")
		assert.Contains(t, rec.Body.String(), "timestamp,ip,user_agent")
	})

	t.Run("success export JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID+"/export?format=json", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "analytics_"+shortID+".json")
		assert.Contains(t, rec.Body.String(), `"ip":"127.0.0.1"`)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID+"/export", nil)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing or invalid Authorization header")
	})

	t.Run("forbidden access by non-owner", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID+"/export", nil)
		req.Header.Set("Authorization", "Bearer "+otherToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Forbidden")
	})

	t.Run("invalid short_id format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/@@@/export", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("unsupported format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/click-count/"+shortID+"/export?format=xml", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		assert.Contains(t, rec.Body.String(), "Unsupported format")
	})
}
