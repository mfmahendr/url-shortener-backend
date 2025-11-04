package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserLinks(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	require.NotNil(t, tcEnv, "tcEnv should be initialized")

	urlSvc := url_service.New(fsService, nil, nil)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)

	// Controller
	controller := controllers.New(urlSvc, nil, fsService, nil)
	controller.Router.GET("/u/shortlinks", authMiddleware.RequireAuth(controller.GetShortlinks))

	// Create test user and token
	userID, token, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "getusershortlinks.user@example.com", nil)
	require.NoError(t, err)

	shortlinkData := []struct {
		ID        string
		URL       string
		IsPrivate string
	}{
		{"u1_pub", "https://public1.example.com", "false"},
		{"u1_prv", "https://private1.example.com", "true"},
		{"u1_pub2", "https://public2.example.com", "false"},
	}

	createdAtNow := time.Now()

	expectedCorrectUserLinksAmout := 0
	for _, s := range shortlinkData {
		isPriv, _ := strconv.ParseBool(s.IsPrivate)
		err := fsService.SetShortlink(ctx, s.ID, models.Shortlink{
			ShortID:   s.ID,
			URL:       s.URL,
			IsPrivate: isPriv,
			CreatedBy: userID,
			CreatedAt: createdAtNow,
		})
		createdAtNow = createdAtNow.Add(time.Second)
		expectedCorrectUserLinksAmout += 1
		require.NoError(t, err)
	}

	// other shortlink created by other.user@example.com to ascertain filtering
	otherUserID, _, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "other.user@example.com", nil)
	require.NoError(t, err)
	err = fsService.SetShortlink(ctx, "other_pub", models.Shortlink{
		ShortID:   "other_pub",
		URL:       "https://otheruser.link",
		CreatedAt: createdAtNow.Add(time.Second),
		CreatedBy: otherUserID,
		IsPrivate: false,
	})
	require.NoError(t, err)

	// TEST CASES
	t.Run("success get all user links (default all)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UserLinksResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, userID, resp.CreatedBy)
		assert.Equal(t, expectedCorrectUserLinksAmout, len(resp.Links))
		for _, l := range resp.Links {
			assert.NotEqual(t, "other_pub", l.ShortID)
		}
	})

	t.Run("success get only private links", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks?is_private=true", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UserLinksResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, userID, resp.CreatedBy)
		assert.NotEmpty(t, resp.Links)
		for _, l := range resp.Links {
			assert.True(t, l.IsPrivate)
			assert.NotEqual(t, "other_pub", l.ShortID)
		}
	})

	t.Run("success get only public links", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks?is_private=false", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UserLinksResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Links)
		for _, l := range resp.Links {
			assert.False(t, l.IsPrivate)
			assert.NotEqual(t, "other_pub", l.ShortID)
		}
	})

	t.Run("pagination works with limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks?limit=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UserLinksResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Links, 1)
		assert.NotEmpty(t, resp.NextCursor)
	})

	t.Run("fail without token (unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks", nil)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("fail with invalid token (unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks", nil)
		req.Header.Set("Authorization", "Bearer invalid_token_here")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("empty response if user has no links", func(t *testing.T) {
		newUserID, newToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "emptyuser@example.com", nil)
		require.NoError(t, err)
		_ = newUserID // not needed explicitly here

		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks", nil)
		req.Header.Set("Authorization", "Bearer "+newToken)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp dto.UserLinksResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Links)
	})

	t.Run("invalid query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/u/shortlinks?is_private=notvalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
