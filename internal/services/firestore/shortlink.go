package firestore_service

import (
	"context"
	"fmt"

	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)

type Shortlink interface {
	DeleteShortlink(ctx context.Context, shortID string) error
	GetShortlink(ctx context.Context, shortID string) (*models.Shortlink, error)
	SetShortlink(ctx context.Context, shortID string, doc models.Shortlink) error
}

func (s *FirestoreServiceImpl) SetShortlink(ctx context.Context, shortID string, doc models.Shortlink) error {
	_, err := s.client.Collection("shortlinks").Doc(shortID).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to set shortlink: %w", err)
	}
	return nil
}

// func (s *FirestoreServiceImpl) GetShortlink(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error) {
func (s *FirestoreServiceImpl) GetShortlink(ctx context.Context, shortID string) (*models.Shortlink, error) {
	docSnap, err := s.client.Collection("shortlinks").Doc(shortID).Get(ctx)
	if !docSnap.Exists() {
		return nil, shortlink_errors.ErrNotFound
	}

	if err != nil {
		return nil, shortlink_errors.ErrFailedRetrieveData
	}

	var shortlink models.Shortlink
	if err := docSnap.DataTo(&shortlink); err != nil {
		return nil, shortlink_errors.ErrFailedRetrieveData
	}

	return &shortlink, nil
}
