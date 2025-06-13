package tracking_service

import (
	"context"
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