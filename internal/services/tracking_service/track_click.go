package tracking_service

import (
	"context"
	"log"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func (t *TrackingServiceImpl) TrackClick(ctx context.Context, shortID, ip, userAgent string) error {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}
	// redis
	if err := t.redis.Incr(ctx, "clicks:"+shortID).Err(); err != nil {
		log.Printf("redis incr failed key=clicks:%s err=%v", shortID, err)
		return err
	}

	clickLog := &models.ClickLog{
		ShortID:   shortID,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	// save to firestore
	if err := t.firestore.AddClickLog(ctx, clickLog); err != nil {
		log.Printf("AddClickLog failed short_id (%s) err: %v", shortID, err)
		return err
	}

	return nil
}
