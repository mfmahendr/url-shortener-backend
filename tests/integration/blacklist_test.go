package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlacklist(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	require.NotNil(t, tcEnv, "tcEnv should be initialized")

	// Middleware + controller setup
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)
	controller := controllers.New(nil, nil, fsService, nil)

	controller.Router.POST("/admin/blacklist", authMiddleware.RequireAdminAuth(controller.AddToBlacklist))
	controller.Router.DELETE("/admin/blacklist", authMiddleware.RequireAdminAuth(controller.RemoveFromBlacklist))
	controller.Router.GET("/admin/blacklist", authMiddleware.RequireAdminAuth(controller.FetchBlacklistItems))

	// Create user and set admin user claim
	claims := map[string]interface{}{
		"admin": true,
	}
	_, token, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "admin@url-shortener.com", &claims)
	require.NoError(t, err)
	_, anotherToken, err := createTestUserAndToken(ctx, authMiddleware.AuthClient, "anotheruser@url-shortener.com", nil)
	require.NoError(t, err)

	t.Run("Successfully blacklist a valid domain", func(t *testing.T) {
		body := `{"type": "domain", "value": "assume-this-as-a-valid-domain-to-be-blacklisted.com"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"added"`)
		assert.Contains(t, rec.Body.String(), `"value":"assume-this-as-a-valid-domain-to-be-blacklisted.com"`)
		assert.Contains(t, rec.Body.String(), `"type":"domain"`)
	})

	t.Run("Successfully blacklist a URL", func(t *testing.T) {
		body := `{"type": "url", "value": "https://malicious.com/phishing"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"added"`)
		assert.Contains(t, rec.Body.String(), `"value":"https://malicious.com/phishing"`)
		assert.Contains(t, rec.Body.String(), `"type":"url"`)
	})

	t.Run("Failed blacklist a domain by a non-admin user", func(t *testing.T) {
		body := `{"domain": "assume-this-as-a-different-valid-domain-to-be-blacklisted.com"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+anotherToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, strings.ToLower(rec.Body.String()), "forbidden")
	})

	t.Run("Failed blacklist an existing domain", func(t *testing.T) {
		body := `{"type": "domain", "value": "assume-this-as-a-valid-domain-to-be-blacklisted.com"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrResourceExists.Error())
	})

	t.Run("Failed blacklist invalid format domain", func(t *testing.T) {
		body := `{"type": "domain", "value": "this-is-not-a-valid-domain!!"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("Add empty domain", func(t *testing.T) {
		body := `{"type": "domain", "value": ""}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Add malformed JSON", func(t *testing.T) {
		body := `{`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Add without token", func(t *testing.T) {
		body := `{"type": "domain", "value": "assume-this-as-another-valid-domain-to-be-blacklisted.com"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/blacklist", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Success remove existing domain", func(t *testing.T) {
		err := fsService.BlacklistDomain(ctx, "a-domain-to-be-removed.com")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/admin/blacklist?type=domain&value=a-domain-to-be-removed.com", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"removed"`)
		assert.Contains(t, rec.Body.String(), `"value":"a-domain-to-be-removed.com"`)
		assert.Contains(t, rec.Body.String(), `"type":"domain"`)
	})

	t.Run("Remove non-existent domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/blacklist?type=domain&value=notfound.com", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrNotFound.Error())
	})

	t.Run("Remove invalid domain format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/blacklist?type=domain&value=!!!invalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), shortlink_errors.ErrValidateRequest.Error())
	})

	t.Run("Fetch blacklisted domains", func(t *testing.T) {
		err := fsService.BlacklistDomain(ctx, "a-fetchable-blacklisted-domain.com")
		assert.NoError(t, err)
		err = fsService.BlacklistDomain(ctx, "another-fetchable-blacklisted-domain.com")
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/admin/blacklist", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var items []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		}

		err = json.Unmarshal(rec.Body.Bytes(), &items)
		require.NoError(t, err)

		values := make([]string, 0)
		for _, item := range items {
			if item.Type == "domain" {
				values = append(values, item.Value)
			}
		}

		assert.Contains(t, values, "a-fetchable-blacklisted-domain.com")
		assert.Contains(t, values, "another-fetchable-blacklisted-domain.com")
	})

	t.Run("Fetch without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/blacklist", nil)
		rec := httptest.NewRecorder()

		controller.Router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
