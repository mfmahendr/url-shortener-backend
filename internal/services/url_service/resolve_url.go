package url_service

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	val "github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func (s *URLServiceImpl) Resolve(ctx context.Context, shortID string) (string, error) {
	if err := val.Validate.Var(shortID, "short_id"); err != nil {
		return "", shortlink_errors.ErrValidateRequest
	}

	shortlink, err := s.shortlink.GetShortlink(ctx, shortID)
	if err != nil {
		return "", err
	}

	// if private, check ownership
	if shortlink.IsPrivate {
		user, ok := ctx.Value(utils.UserKey).(string)
		if !ok || user == "" || user != shortlink.CreatedBy {
			return "", shortlink_errors.ErrForbidden
		}
	}

	return shortlink.URL, nil
}