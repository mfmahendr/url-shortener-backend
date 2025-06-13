package tracking_service

import (
	"context"


	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/redis/go-redis/v9"
)

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