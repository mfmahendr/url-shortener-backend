package tracking_service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

// *--- MOCK DEFINITIONS ---* //
// Firestore ClickLog SERVICE
type MockClickLogStore struct{ mock.Mock }

func (m *MockClickLogStore) AddClickLog(ctx context.Context, log *models.ClickLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockClickLogStore) GetClickLogs(ctx context.Context, query dto.ClickLogsQuery) ([]models.ClickLog, string, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]models.ClickLog), args.String(1), args.Error(2)
}

func (m *MockClickLogStore) StreamClickLogs(ctx context.Context, shortID string) (*firestore.DocumentIterator, error) {
	args := m.Called(ctx, shortID)
	return args.Get(0).(*firestore.DocumentIterator), args.Error(1)
}

func (m *MockClickLogStore) GetAnalytics(ctx context.Context, shortID string) (int64, []models.ClickLog, error) {
	args := m.Called(ctx, shortID)
	return args.Get(0).(int64), args.Get(1).([]models.ClickLog), args.Error(2)
}

func TestMain(m *testing.M) {
	validators.Init()
	os.Exit(m.Run())
}

// *--- TEST CASES ---* //
func TestTrackingService_TrackClick(t *testing.T) {
	db, redisMock := redismock.NewClientMock()
	store := new(MockClickLogStore)
	svc := tracking_service.New(store, db)

	t.Run("Success", func(t *testing.T) {
		shortID := "abc123"
		ctx := context.Background()

		redisMock.ExpectIncr("clicks:" + shortID).SetVal(1)
		store.On("AddClickLog", ctx, mock.Anything).Return(nil)

		err := svc.TrackClick(ctx, shortID, "127.0.0.1", "Mozilla")
		assert.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("Invalid ShortID", func(t *testing.T) {
		err := svc.TrackClick(context.Background(), "", "127.0.0.1", "Mozilla")
		require.Error(t, err)
		assert.Equal(t, shortlink_errors.ErrValidateRequest, err)
	})
}

func TestTrackingService_GetClickCount(t *testing.T) {
	redisClient, clientMock := redismock.NewClientMock()
	store := new(MockClickLogStore)
	svc := tracking_service.New(store, redisClient)

	ctx := context.Background()
	shortID := "abc123"

	t.Run("Success - With Data", func(t *testing.T) {
		clientMock.ExpectGet("clicks:" + shortID).SetVal("5")

		count, err := svc.GetClickCount(ctx, shortID)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
		assert.NoError(t, clientMock.ExpectationsWereMet())
	})

	t.Run("Success - No Data (Redis.Nil)", func(t *testing.T) {
		clientMock.ExpectGet("clicks:missing").RedisNil()

		count, err := svc.GetClickCount(ctx, "missing")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.NoError(t, clientMock.ExpectationsWereMet())
	})

	t.Run("Invalid ShortID", func(t *testing.T) {
		count, err := svc.GetClickCount(ctx, "")
		require.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.Equal(t, shortlink_errors.ErrValidateRequest, err)
	})
}

func TestTrackingService_GetAnalytics(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	store := new(MockClickLogStore)
	svc := tracking_service.New(store, rdb)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		query := dto.ClickLogsQuery{ShortID: "correctshortid"}
		logs := []models.ClickLog{
			{Timestamp: time.Now(), IP: ":abcd:1", UserAgent: "A User Agent"},
			{Timestamp: time.Now(), IP: "127.0.0.1", UserAgent: "Another UA"},
		}

		store.On("GetClickLogs", mock.Anything, query).Return(logs, "", nil)

		result, err := svc.GetAnalytics(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(len(logs)), result.TotalClicks)
		assert.Equal(t, query.ShortID, result.ShortID)
		store.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		query := dto.ClickLogsQuery{}
		result, err := svc.GetAnalytics(ctx, query)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, shortlink_errors.ErrValidateRequest, err)
	})

	t.Run("Store Error", func(t *testing.T) {
		query := dto.ClickLogsQuery{ShortID: "incorrectshortid"}		// make sure this value is different from Success test case
		store.On("GetClickLogs", mock.Anything, query).Return([]models.ClickLog(nil), "", shortlink_errors.ErrFailedRetrieveData)

		result, err := svc.GetAnalytics(ctx, query)
		require.Error(t, err)
		assert.Nil(t, result)

		store.AssertExpectations(t)
	})

}
