package firestore_service

import (
	"context"
	"fmt"
	
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

type FirestoreService interface {
	GetClient() *firestore.Client 
}

type FirestoreServiceImpl struct {
	client *firestore.Client
}

func New(ctx context.Context, firebaseApp *firebase.App) (*FirestoreServiceImpl, error) {
	client, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firestore client: %w", err)
	}
	return &FirestoreServiceImpl{client: client}, nil
}

func (s *FirestoreServiceImpl) GetClient() *firestore.Client {
	return s.client
}
