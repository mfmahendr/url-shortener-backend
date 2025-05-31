package url_service

import (
	"context"
	"errors"
	"log"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	val "github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

type URLService interface {
	Shorten(ctx context.Context, req dto.ShortenerRequest) (shortID string, err error)
	Resolve(ctx context.Context, shortID string) (string, error)
}

type URLServiceImpl struct {
	firestore firestore_service.FirestoreService
}

func New(firestoreService firestore_service.FirestoreService) URLService {
	return &URLServiceImpl{firestore: firestoreService}
}

func (s *URLServiceImpl) Shorten(ctx context.Context, req dto.ShortenerRequest) (shortID string, err error) {
	if err := val.Validate.Struct(req); err != nil {
		return "", shortlink_errors.ErrValidateRequest
	}

	if req.CustomID == "" {
		// generate short id with nanoid
		shortID, err = nanoid.New()
		if err != nil {
			log.Println("Error generating ID:", err)
			return "", shortlink_errors.ErrGenerateID
		}
	} else {
		//  blacklist some keywords
		if utils.BlacklistedCustomIDs[req.CustomID] {
			return "", shortlink_errors.ErrBlacklistedID
		}

		// check exists
		doc, err := s.firestore.GetShortlink(ctx, req.CustomID)
		if err == nil && doc.Exists() {
			return "", shortlink_errors.ErrIDExists
		}
		shortID = req.CustomID
	}

	doc := map[string]interface{}{
		"short_id":   shortID,
		"url":        req.URL,
		"created_at": time.Now(),
	}

	error := s.firestore.SetShortlink(ctx, shortID, doc)
	if error != nil {
		return "", shortlink_errors.ErrSaveShortlink
	}

	return shortID, error
}

func (s *URLServiceImpl) Resolve(ctx context.Context, shortID string) (string, error) {
	if err := val.Validate.Var(shortID, "short_id"); err != nil {
		return "", shortlink_errors.ErrValidateRequest
	}

	doc, err := s.firestore.GetShortlink(ctx, shortID)
	if err != nil || !doc.Exists() {
		return "", errors.New("not found")
	}
	url, _ := doc.DataAt("url")
	return url.(string), nil
}
