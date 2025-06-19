package integration

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirect(t *testing.T) {
	ctx := context.Background()
	tcEnv = GetSharedTestContainerEnv(ctx, t)
	if tcEnv == nil {
		t.Fatal("tcEnv is nil! initialization likely failed")
	}
	validators.Init()

	// Service dependencies
	urlSvc := url_service.New(fsService, fsService, nil)
	trackingSvc := tracking_service.New(fsService, tcEnv.rdClient)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(tcEnv.FsApp)

	rateLimiter := middleware.NewRateLimiter(tcEnv.rdClient)
	rateLimiter.SetLimit(5, 5*time.Second)

	// Controllers
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
	privateOwnerUID, shortIDOwnerToken := createTestUserAndToken(t, "owner@email.com")
	_, anotherToken := createTestUserAndToken(t, "anotheruser@email.com")


	// create shortlink to redirect
	err := fsService.SetShortlink(ctx, existingPublicShortID, models.Shortlink{
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

func createTestUserAndToken(t *testing.T, email string) (string, string) {
	body := map[string]interface{}{
		"email":             email,
		"password":          "password123",
		"returnSecureToken": true,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(
		"http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake-api-key",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var res struct {
		LocalID string `json:"localId"`		// UID
		IDToken string `json:"idToken"`		// Token
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

	return res.LocalID, res.IDToken
}
