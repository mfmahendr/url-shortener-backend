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

func TestRedirect(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)

	// middleware
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)

	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(5, 5*time.Second)

	// service and controllers
	urlSvc := url_service.New(fsService, fsService, nil)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)
	controller := controllers.New(urlSvc, trackingSvc, fsService, nil)
	controller.Router.GET("/r/:short_id",
		rateLimiter.Apply(
			authMiddleware.OptionalAuth(controller.Redirect),
		),
	)

	// Prepare test data
	existingPublicShortID := "public123"
	privateShortID := "private123"
	expectedURL := "https://google.com"		// both short id will redirect to this URL

	// creating auth user
	privateOwnerUID, shortIDOwnerToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "owner@email.com", nil)
	require.NoError(t, err)
	_, anotherToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "anotheruser@email.com", nil)
	require.NoError(t, err)

	// create shortlink to redirect
	err = fsService.SetShortlink(ctx, existingPublicShortID, models.Shortlink{
		ShortID:   existingPublicShortID,
		URL:       expectedURL,
		IsPrivate: false,
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	err = fsService.SetShortlink(ctx, privateShortID, models.Shortlink{
		ShortID:   privateShortID,
		URL:       expectedURL,
		IsPrivate: true,
		CreatedBy: privateOwnerUID,
		CreatedAt: time.Now(),
	})
	assert.NoError(t, err)

	t.Run("success redirect to public id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/"+existingPublicShortID, nil)
		req.RemoteAddr = "192.0.2.1:12345"

		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, expectedURL, rec.Header().Get("Location"))
	})

	t.Run("success redirect from private ID by the owner", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/"+privateShortID, nil)
		req.Header.Set("Authorization", "Bearer "+shortIDOwnerToken)
		req.RemoteAddr = "192.0.2.1:12345"

		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, expectedURL, rec.Header().Get("Location"))
	})

	t.Run("forbidden redirect from private ID by another user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/"+privateShortID, nil)
		req.Header.Set("Authorization", "Bearer "+anotherToken)
		req.RemoteAddr = "192.0.2.1:12345"

		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrForbidden.Error())
	})

	t.Run("not found shortlink", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/notexist", nil)
		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrNotFound.Error())
	})

	t.Run("invalid short ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/@@@", nil)
		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("success redirect with forwarded IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/"+existingPublicShortID, nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195") // example X-Forwarded-For client
		req.RemoteAddr = "192.0.2.1:12345"

		rec := httptest.NewRecorder()
		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, expectedURL, rec.Header().Get("Location"))
	})
}