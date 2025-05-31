package url_service

import (
	"context"
	"log"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	val "github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

type URLService interface {
	Shorten(ctx context.Context, req dto.ShortenRequest) (shortID string, err error)
	Resolve(ctx context.Context, shortID string) (string, error)
	IsOwner(ctx context.Context, shortID string, uid string) (bool, error)
}

type URLServiceImpl struct {
	firestore firestore_service.Shortlink
}

func New(shortlinkService firestore_service.Shortlink) URLService {
	return &URLServiceImpl{firestore: shortlinkService}
}

func (s *URLServiceImpl) Shorten(ctx context.Context, req dto.ShortenRequest) (shortID string, err error) {
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
		_, err := s.firestore.GetShortlink(ctx, req.CustomID)
		if err != nil && err != shortlink_errors.ErrNotFound {
			return "", shortlink_errors.ErrIDExists
		}
		shortID = req.CustomID
	}

	user := ctx.Value("uid")
	doc := models.Shortlink{
		ShortID:   shortID,
		URL:       req.URL,
		CreatedAt: time.Now(),
		CreatedBy: user.(string),
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

	shortlink, err := s.firestore.GetShortlink(ctx, shortID)
	if err != nil {
		return "", err
	}

	return shortlink.URL, nil
}

func (s *URLServiceImpl) IsOwner(ctx context.Context, shortID string, uid string) (bool, error) {
	shortlink, err := s.firestore.GetShortlink(ctx, shortID)
	if err != nil {
		return false, err
	}

	createdBy := shortlink.CreatedBy
	return createdBy == uid, nil
}
