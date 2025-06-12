package url_service

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/sync/errgroup"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	val "github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func (s *URLServiceImpl) Shorten(ctx context.Context, req dto.ShortenRequest) (string, error) {
	if err := val.Validate.Struct(req); err != nil {
		return "", shortlink_errors.ErrValidateRequest
	}

	// if CustomID is not provided, generate a new ID
	if req.CustomID == "" {
		id, err := nanoid.New()
		if err != nil {
			log.Println("Error generating ID:", err)
			return "", shortlink_errors.ErrGenerateID
		}
		return s.saveShortlink(ctx, id, req.URL)
	}

	if err := s.validateCustomID(ctx, req); err != nil {
		return "", err
	}

	return s.saveShortlink(ctx, req.CustomID, req.URL)
}

func (s *URLServiceImpl) validateCustomID(ctx context.Context, req dto.ShortenRequest) error {
	if utils.BlacklistedCustomIDs[req.CustomID] { // check reserved keywords
		return shortlink_errors.ErrBlacklistedID
	}

	// check if the URL is valid
	parsedURL, err := url.Parse(req.URL)
	if err != nil || parsedURL.Host == "" {
		log.Println("Invalid request: URL has no host")
		return shortlink_errors.ErrValidateRequest
	}

	eg, ctx := errgroup.WithContext(ctx)
	domain := strings.ToLower(parsedURL.Hostname())

	// check custom blacklisted domain
	eg.Go(func() error {
		isBlacklisted, err := s.blacklist.IsDomainBlacklisted(ctx, domain)
		if err != nil {
			log.Println("Failed to retrieve data")
			return shortlink_errors.ErrFailedRetrieveData
		}
		if isBlacklisted {
			log.Println("This domain is blacklisted")
			return shortlink_errors.ErrBlacklistedID
		}
		return nil
	})

	// check safe Browsing
	eg.Go(func() error {
		isUnsafe, err := s.safebrowsing.IsUnsafe(ctx, req.URL)
		if err != nil {
			log.Printf("SafeBrowsing error: %v", err)
			return shortlink_errors.ErrFailedRetrieveData
		}
		if isUnsafe {
			log.Println("This site is unsafe")
			return shortlink_errors.ErrValidateRequest
		}
		return nil
	})

	// check custom ID existence
	eg.Go(func() error {
		_, err := s.shortlink.GetShortlink(ctx, req.CustomID)
		if err != nil && !errors.Is(err, shortlink_errors.ErrNotFound) {
			return shortlink_errors.ErrIDExists
		}
		return nil
	})

	return eg.Wait()
}

// Simpan shortlink (reusable function)
func (s *URLServiceImpl) saveShortlink(ctx context.Context, shortID, url string) (string, error) {
	user, ok := ctx.Value(utils.UserKey).(string)
	if !ok {
		return "", shortlink_errors.ErrValidateRequest
	}

	doc := models.Shortlink{
		ShortID:   shortID,
		URL:       url,
		CreatedAt: time.Now(),
		CreatedBy: user,
	}

	if err := s.shortlink.SetShortlink(ctx, shortID, doc); err != nil {
		return "", shortlink_errors.ErrSaveShortlink
	}
	return shortID, nil
}
