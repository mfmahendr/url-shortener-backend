package firestore_service

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type ClickLog interface {
	AddClickLog(ctx context.Context, doc interface{}) error
	GetClickLog(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error)
}

func (s *FirestoreServiceImpl) AddClickLog(ctx context.Context, doc interface{}) error {
	_, _, err := s.client.Collection("click_logs").Add(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to add click_logs: %w", err)
	}
	return nil
}

func (s *FirestoreServiceImpl) GetClickLog(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error) {
	docSnap, err := s.client.Collection("click_logs").Doc(shortID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get click log: %w", err)
	}
	return docSnap, nil
}