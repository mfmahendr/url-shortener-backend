package url_service_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

// *--- MOCK DEFINITIONS ---* //
// Firestore shortlink SERVICE
type MockShortlink struct{ mock.Mock }

func (m *MockShortlink) SetShortlink(ctx context.Context, shortID string, link models.Shortlink) error {
	args := m.Called(ctx, shortID, link)
	return args.Error(0)
}

func (m *MockShortlink) GetShortlink(ctx context.Context, shortID string) (*models.Shortlink, error) {
	args := m.Called(ctx, shortID)
	return args.Get(0).(*models.Shortlink), args.Error(1)
}

// Firestore blacklist checker SERVICE
type MockBlacklistChecker struct{ mock.Mock }

func (m *MockBlacklistChecker) IsDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	args := m.Called(ctx, domain)
	return args.Bool(0), args.Error(1)
}

// Safebrowsing safety checker SERVICE
type MockURLSafetyChecker struct{ mock.Mock }

func (m *MockURLSafetyChecker) IsUnsafe(ctx context.Context, url string) (bool, error) {
	args := m.Called(ctx, url)
	return args.Bool(0), args.Error(1)
}

func TestMain(m *testing.M) {
	validators.Init()
	os.Exit(m.Run())
}

// *--- TEST CASES ---* //
// Check Owner
func TestIsOwner(t *testing.T) {
	mockSL := new(MockShortlink)
	mockBL := new(MockBlacklistChecker)
	mockSB := new(MockURLSafetyChecker)
	svc := url_service.New(mockSL, mockBL, mockSB)

	t.Run("IsOwner returns true when user is owner", func(t *testing.T) {
		shortID := "test123"
		uid := "user123"
		mockSL.On("GetShortlink", mock.Anything, shortID).Return(&models.Shortlink{
			ShortID:   shortID,
			CreatedBy: uid,
		}, nil).Once()

		isOwner, err := svc.IsOwner(context.Background(), shortID, uid)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})

	t.Run("IsOwner returns false when user is not the owner", func(t *testing.T) {
		shortID := "test123"
		uid := "otherUser"
		mockSL.On("GetShortlink", mock.Anything, shortID).Return(&models.Shortlink{
			ShortID:   shortID,
			CreatedBy: "ownerUser",
		}, nil).Once()

		isOwner, err := svc.IsOwner(context.Background(), shortID, uid)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})

	t.Run("IsOwner returns error when shortlink not found", func(t *testing.T) {
		shortID := "nonexistent"
		mockSL.On("GetShortlink", mock.Anything, shortID).
			Return(&models.Shortlink{}, shortlink_errors.ErrNotFound).Once()

		isOwner, err := svc.IsOwner(context.Background(), shortID, "user")
		assert.Error(t, err)
		assert.False(t, isOwner)
		assert.Equal(t, shortlink_errors.ErrNotFound, err)
	})
}

func TestResolve(t *testing.T) {
	mockSL := new(MockShortlink)
	mockBL := new(MockBlacklistChecker)
	mockSB := new(MockURLSafetyChecker)
	svc := url_service.New(mockSL, mockBL, mockSB)

	t.Run("Public URL resolves successfully", func(t *testing.T) {
		shortID := "abc123"
		mockSL.On("GetShortlink", mock.Anything, shortID).Return(&models.Shortlink{
			ShortID:   shortID,
			URL:       "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url",
			IsPrivate: false,
			CreatedBy: "user123",
		}, nil).Once()

		url, err := svc.Resolve(context.Background(), shortID)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url", url)
	})

	t.Run("Private URL resolves if user is owner", func(t *testing.T) {
		shortID := "abc123"
		ctx := context.WithValue(context.Background(), utils.UserKey, "user123")

		mockSL.On("GetShortlink", mock.Anything, shortID).Return(&models.Shortlink{
			ShortID:   shortID,
			URL:       "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url",
			IsPrivate: true,
			CreatedBy: "user123",
		}, nil).Once()

		url, err := svc.Resolve(ctx, shortID)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url", url)
	})

	t.Run("Private URL returns forbidden for other users", func(t *testing.T) {
		shortID := "abc123"
		ctx := context.WithValue(context.Background(), utils.UserKey, "differentuser")

		mockSL.On("GetShortlink", mock.Anything, shortID).Return(&models.Shortlink{
			ShortID:   shortID,
			URL:       "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url",
			IsPrivate: true,
			CreatedBy: "user123",
		}, nil).Once()

		url, err := svc.Resolve(ctx, shortID)
		assert.Error(t, err)
		assert.Equal(t, "", url)
		assert.Equal(t, shortlink_errors.ErrForbidden, err)
	})
}

// Shorten
func TestShorten_SuccessWithCustomID(t *testing.T) {
	mockSL := new(MockShortlink)
	mockBL := new(MockBlacklistChecker)
	mockSB := new(MockURLSafetyChecker)

	svc := url_service.New(mockSL, mockBL, mockSB)
	t.Run("URL with Custom ID has successfully shortened", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), utils.UserKey, "user123")
		req := dto.ShortenRequest{
			URL:       "https://example.com?id=this-is-supposed-to-be-an-extremely-extra-long-url",
			CustomID:  "custom123",
			IsPrivate: true,
		}

		mockBL.On("IsDomainBlacklisted", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx.Value(utils.UserKey) == "user123"
		}), "example.com").Return(false, nil) 										// domain is not blacklisted
		mockSB.On("IsUnsafe", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx.Value(utils.UserKey) == "user123"
		}), req.URL).Return(false, nil) 											// domain is safe
		mockSL.On("GetShortlink", mock.MatchedBy(func(c context.Context) bool {
			return c.Value(utils.UserKey) == "user123"
		}), req.CustomID).Return(&models.Shortlink{}, shortlink_errors.ErrNotFound) // shortlink not already in the database

		mockSL.On("SetShortlink", mock.MatchedBy(func(c context.Context) bool {
			return c.Value(utils.UserKey) == "user123"
		}), req.CustomID, mock.Anything).Return(nil) 								// successful set and save the shortlink

		shortID, err := svc.Shorten(ctx, req)

		require.NoError(t, err)
		require.Equal(t, req.CustomID, shortID)
	})
}

func TestShorten_InvalidURL(t *testing.T) {
	mockSL := new(MockShortlink)
	mockBL := new(MockBlacklistChecker)
	mockSB := new(MockURLSafetyChecker)

	svc := url_service.New(mockSL, mockBL, mockSB)
	t.Run("Invalid URL failed to be shortened", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), utils.UserKey, "user123")
		req := dto.ShortenRequest{
			URL:      "this-is-a-bad/invalid-url:/",
			CustomID: "abc123",
		}

		shortID, err := svc.Shorten(ctx, req)
		require.Equal(t, "", shortID)
		require.Error(t, err)
		require.Equal(t, shortlink_errors.ErrValidateRequest, err)
	})
}
