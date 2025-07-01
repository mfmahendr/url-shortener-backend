package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	safebrowsing_service "github.com/mfmahendr/url-shortener-backend/internal/services/safebrowsing"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShorten(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)

	// Init services
	mockSB := &safebrowsing_service.MockSafeBrowsingService{
		UnsafeURLs: map[string]bool{
			"http://malware.testing.google.test/testing/malware/": true,
		},
	}

	urlSvc := url_service.New(fsService, fsService, mockSB)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)

	// Controller
	controller := controllers.New(urlSvc, trackingSvc, fsService, nil)
	controller.Router.POST("/u/shorten",
		authMiddleware.RequireAuth(controller.Shorten),
	)

	// Create test user and token
	userID, token, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "shortenuser@example.com", nil)
	require.NoError(t, err)

	// add blacklisted domain data
	err = fsService.BlacklistDomain(ctx, "this-is-a-blacklisted-domain.com")
	require.NoError(t, err)

	customID := "customtest123"
	var existingCustomID string

	t.Run("success shorten with generated ID", func(t *testing.T) {
		body := map[string]interface{}{
			"url":        "https://this-custom-id-must-be-successfully-shorten.com",
			"is_private": false,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res dto.ShortenResponse
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.NotEmpty(t, res.ShortID)

		expectedJSON := fmt.Sprintf(`{"short_id": "%s"}`, res.ShortID)
		assert.JSONEq(t, expectedJSON, rec.Body.String())

		// Verify firestore data
		shortlink, err := fsService.GetShortlink(ctx, res.ShortID)
		require.NoError(t, err)
		assert.Equal(t, "https://this-custom-id-must-be-successfully-shorten.com", shortlink.URL)
		assert.Equal(t, userID, shortlink.CreatedBy)
		assert.False(t, shortlink.IsPrivate)
	})

	t.Run("success shorten with custom ID", func(t *testing.T) {
		body := map[string]interface{}{
			"url":        "https://www.success-shorten-with-custom-id.com",
			"custom_id":  customID,
			"is_private": true,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		expectedJSON := fmt.Sprintf(`{"short_id": "%s"}`, customID)
		assert.JSONEq(t, expectedJSON, rec.Body.String())

		// Verify firestore data
		shortlink, err := fsService.GetShortlink(ctx, customID)
		require.NoError(t, err)
		assert.Equal(t, "https://www.success-shorten-with-custom-id.com", shortlink.URL)
		assert.Equal(t, userID, shortlink.CreatedBy)
		assert.True(t, shortlink.IsPrivate)

		existingCustomID = customID		// for existing customID test case
	})

	t.Run("fail without token", func(t *testing.T) {
		body := map[string]interface{}{
			"url": "https://this-is-unauthorized-and-will-fail-without-token.com",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing or invalid Authorization header")
	})

	t.Run("fail with invalid token", func(t *testing.T) {
		body := map[string]interface{}{
			"url": "https://unauthorized-unauthorized-unauthorized-unauthorized.com",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer invalid_token")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid token")
	})

	t.Run("fail without URL", func(t *testing.T) {
		body := map[string]interface{}{}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("fail with invalid URL", func(t *testing.T) {
		body := map[string]interface{}{
			"url": "inv@@@lid-url",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("fail with reserved custom ID keyword", func(t *testing.T) {
		body := map[string]interface{}{
			"url":       "https://www.facebook.com",
			"custom_id": "admin",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrBlacklistedID.Error())
	})

	t.Run("fail with existing custom ID", func(t *testing.T) {
		body := map[string]interface{}{
			"url":       "https://www.instagram.com",
			"custom_id": existingCustomID, // already be added
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrIDExists.Error())
	})

	t.Run("fail with blacklisted domain", func(t *testing.T) {
		body := map[string]interface{}{
			"url": "https://this-is-a-blacklisted-domain.com?query=must&be=faield",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrForbiddenInput.Error())
	})

	t.Run("fail with unsafe domain (safebrowsing)", func(t *testing.T) {
		body := map[string]interface{}{
			"url": "http://malware.testing.google.test/testing/malware/",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/u/shorten", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrForbiddenInput.Error())
	})
}
