package url_service

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func (s *URLServiceImpl) IsOwner(ctx context.Context, shortID string, uid string) (bool, error) {
	if err := validators.Validate.Var(shortID, "short_id"); err != nil {
		return false, shortlink_errors.ErrValidateRequest
	}

	shortlink, err := s.shortlink.GetShortlink(ctx, shortID)
	if err != nil {
		return false, err
	}

	createdBy := shortlink.CreatedBy
	return createdBy == uid, nil
}
