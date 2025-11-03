package tracking_service

import (
	"context"
	"net/http"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	firestoreService "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/redis/go-redis/v9"
)

type (
	TrackingService interface {
		GetClickCount(ctx context.Context, shortID string) (int64, error)
		TrackClick(ctx context.Context, shortID, ip, userAgent string) error
		StreamClickLogs(ctx context.Context, w http.ResponseWriter, req dto.ClickLogsRequest) error
		GetAnalytics(ctx context.Context, req dto.ClickLogsRequest) (*dto.AnalyticsDTO, error)
	}

	TrackingServiceImpl struct {
		firestore firestoreService.ClickLog
		redis     *redis.Client
	}
)

func New(fs firestoreService.ClickLog, redis *redis.Client) TrackingService {
	return &TrackingServiceImpl{firestore: fs, redis: redis}
}