package url_service

import (
	"context"
	"errors"
	"log"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"

	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)

type URLService struct {
	Firestore *firestore_service.FirestoreService
}

func New(firestoreService *firestore_service.FirestoreService) *URLService {
	return &URLService{Firestore: firestoreService}
}

func (s *URLService) Shorten(ctx context.Context, url, customID string) (shortID string, err error) {
	if customID == "" {
		// generate short id with nanoid
		shortID, err = nanoid.New()
		if err != nil {
			log.Println("Error generating ID:", err)
			return "", shortlink_errors.ErrGenerateID
		}
	} else {
		// check exists
		doc, err := s.Firestore.GetShortlink(ctx, customID)
		if err == nil && doc.Exists() {
			return "", shortlink_errors.ErrIDExists
		}
		shortID = customID
	}

	doc := map[string]interface{}{
		"short_id":   shortID,
		"url":        url,
		"created_at": time.Now(),
	}

	error := s.Firestore.SetShortlink(ctx, shortID, doc)
	if error != nil {
		return "", shortlink_errors.ErrSaveShortlink
	}

	return shortID, error
}

func (s *URLService) Resolve(ctx context.Context, shortID string) (string, error) {
	doc, err := s.Firestore.GetShortlink(ctx, shortID)
	if err != nil || !doc.Exists() {
		return "", errors.New("not found")
	}
	url, _ := doc.DataAt("url")
	return url.(string), nil
}
