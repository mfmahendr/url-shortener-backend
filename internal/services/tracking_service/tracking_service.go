package tracking_service

import (
	"context"
	"log"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	firestore "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/redis/go-redis/v9"
)

type TrackingService interface {
	GetClickCount(ctx context.Context, shortID string) (int64, error)
	TrackClick(ctx context.Context, shortID, ip, userAgent string) error
	GetAnalytics(ctx context.Context, query dto.ClickLogsQuery) (*dto.AnalyticsDTO, error)
}

type TrackingServiceImpl struct {
	firestore firestore.ClickLog
	redis     *redis.Client
}

func New(fs firestore.ClickLog, redis *redis.Client) TrackingService {
	return &TrackingServiceImpl{firestore: fs, redis: redis}
}

func (t *TrackingServiceImpl) TrackClick(ctx context.Context, shortID, ip, userAgent string) error {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}
	// redis
	if err := t.redis.Incr(ctx, "clicks:"+shortID).Err(); err != nil {
		return err
	}

	clickLog := &models.ClickLog{
		ShortID:   shortID,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	// save to firestore
	return t.firestore.AddClickLog(ctx, clickLog)
}

func (t *TrackingServiceImpl) GetClickCount(ctx context.Context, shortID string) (int64, error) {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return 0, shortlink_errors.ErrValidateRequest
	}

	count, err := t.redis.Get(ctx, "clicks:"+shortID).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, shortlink_errors.ErrFailedRetrieveData
	}

	return count, nil
}

func (t *TrackingServiceImpl) GetAnalytics(ctx context.Context, query dto.ClickLogsQuery) (*dto.AnalyticsDTO, error) {
	if err := validators.Validate.Struct(query); err != nil {
		log.Println("Validation error:", err)
		return nil, shortlink_errors.ErrValidateRequest
	}

	logs, nextCursor, err := t.firestore.GetClickLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	dtoLogs := make([]dto.ClickLogDTO, 0, len(logs))
	var count int64 = 0
	for _, l := range logs {
		dtoLogs = append(dtoLogs, dto.ClickLogDTO{
			Timestamp: l.Timestamp,
			IP:        l.IP,
			UserAgent: l.UserAgent,
		})
		count++
	}

	responseData := &dto.AnalyticsDTO{
		ShortID:     query.ShortID,
		TotalClicks: count,
		Clicks:      dtoLogs,
		NextCursor:  nextCursor,
	}

	return responseData, nil
}
