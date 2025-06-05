package url_service

import (
	"context"
)

func (s *URLServiceImpl) IsOwner(ctx context.Context, shortID string, uid string) (bool, error) {
	shortlink, err := s.shortlink.GetShortlink(ctx, shortID)
	if err != nil {
		return false, err
	}

	createdBy := shortlink.CreatedBy
	return createdBy == uid, nil
}
