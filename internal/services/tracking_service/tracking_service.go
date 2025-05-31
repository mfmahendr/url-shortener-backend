package tracking_service

import (
	"context"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	firestore "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/redis/go-redis/v9"
)

type TrackingService struct {
	Firestore firestore.ClickLog
	Redis     *redis.Client
}

func New(fs firestore.ClickLog, redis *redis.Client) *TrackingService {
	return &TrackingService{Firestore: fs, Redis: redis}
}

func (t *TrackingService) TrackClick(ctx context.Context, shortID, ip, userAgent string) error {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}
	// redis
	if err := t.Redis.Incr(ctx, "clicks:"+shortID).Err(); err != nil {
		return err
	}

	clickLog := &models.ClickLog{
		ShortID:    shortID,
		IP:         ip,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	// save to firestore
	return t.Firestore.AddClickLog(ctx, clickLog)
}

func (t *TrackingService) GetClickCount(ctx context.Context, shortID string) (int64, error) {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return 0, shortlink_errors.ErrValidateRequest
	}

	count, err := t.Redis.Get(ctx, "clicks:"+shortID).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, shortlink_errors.ErrFailedRetrieveData
	}

	return count, nil
}

func (t *TrackingService) GetAnalytics(ctx context.Context, shortID string) (*dto.AnalyticsDTO, error) {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return nil, shortlink_errors.ErrValidateRequest
	}

	count, logs, err := t.Firestore.GetAnalytics(ctx, shortID)
	if err != nil {
		return nil, err
	}

	responseData := &dto.AnalyticsDTO{
		ShortID:     shortID,
		TotalClicks: count,
		Clicks:      logs,
	}

	return responseData, nil
}
