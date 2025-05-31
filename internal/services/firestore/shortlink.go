package firestore_service

import (
	"context"
	"fmt"
	
	"cloud.google.com/go/firestore"
)

type Shortlink interface {
	DeleteShortlink(ctx context.Context, shortID string) error
	GetShortlink(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error)
	SetShortlink(ctx context.Context, shortID string, doc interface{}) error
}

func (s *FirestoreServiceImpl) SetShortlink(ctx context.Context, shortID string, doc interface{}) error {
	_, err := s.client.Collection("shortlinks").Doc(shortID).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to set shortlink: %w", err)
	}
	return nil
}

func (s *FirestoreServiceImpl) GetShortlink(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error) {
	docSnap, err := s.client.Collection("shortlinks").Doc(shortID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get shortlink: %w", err)
	}
	return docSnap, nil
}

func (s *FirestoreServiceImpl) DeleteShortlink(ctx context.Context, shortID string) error {
	_, err := s.client.Collection("shortlinks").Doc(shortID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete shortlink: %w", err)
	}
	return nil
}
