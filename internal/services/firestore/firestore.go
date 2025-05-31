package firestore_service

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

type FirestoreService struct {
	client *firestore.Client
}

func New(ctx context.Context, firebaseApp *firebase.App) (*FirestoreService, error) {
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firestore client: %w", err)
	}
	return &FirestoreService{client: client}, nil
}

func (s *FirestoreService) GetClient() *firestore.Client {
	return s.client
}

func (s *FirestoreService) SetShortlink(ctx context.Context, shortID string, doc interface{}) error {
	_, err := s.client.Collection("shortlinks").Doc(shortID).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to set shortlink: %w", err)
	}
	return nil
}

func (s *FirestoreService) GetShortlink(ctx context.Context, shortID string) (*firestore.DocumentSnapshot, error) {
	docSnap, err := s.client.Collection("shortlinks").Doc(shortID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get shortlink: %w", err)
	}
	return docSnap, nil
}

func (s *FirestoreService) DeleteShortlink(ctx context.Context, shortID string) error {
	_, err := s.client.Collection("shortlinks").Doc(shortID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete shortlink: %w", err)
	}
	return nil
}
