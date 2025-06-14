package safebrowsing_service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redismock "github.com/go-redis/redismock/v9"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

// *--- MOCK DEFINITION ---* //
// Helper
type testSafeBrowsingService struct {
	SafeBrowsingServiceImpl
	MockedRequest func(url string) (bool, error)
}

func TestMain(m *testing.M) {
	validators.Init()
	os.Exit(m.Run())
}

// --- Tests --- //
func TestSafeBrowsingService_CacheHitSafe(t *testing.T) {
	ctx := context.Background()
	redisClient, mock := redismock.NewClientMock()

	const testURL = "http://example.com?assume=this-is-a-dangerous-malicious-link"
	cacheKey := "safebrowsing:" + testURL

	t.Run("Invalid URL", func(t *testing.T) {
		service := &testSafeBrowsingService{}
		isUnsafe, err := service.IsUnsafe(ctx, "invalid-url")
		require.Error(t, err)
		assert.False(t, isUnsafe)
		assert.Equal(t, shortlink_errors.ErrValidateRequest, err)
	})

	t.Run("Cache Hit - Safe", func(t *testing.T) {
		mock.ExpectGet(cacheKey).SetVal("safe")

		service := &testSafeBrowsingService{SafeBrowsingServiceImpl: SafeBrowsingServiceImpl{
			redis: redisClient,
		}}

		isUnsafe, err := service.IsUnsafe(ctx, testURL)
		require.NoError(t, err)
		assert.False(t, isUnsafe)
	})

	t.Run("Cache Hit - Unsafe", func(t *testing.T) {
		mock.ExpectGet(cacheKey).SetVal("unsafe")

		service := &testSafeBrowsingService{SafeBrowsingServiceImpl: SafeBrowsingServiceImpl{
			redis: redisClient,
		}}

		isUnsafe, err := service.IsUnsafe(ctx, testURL)
		require.NoError(t, err)
		assert.True(t, isUnsafe)
	})

}
