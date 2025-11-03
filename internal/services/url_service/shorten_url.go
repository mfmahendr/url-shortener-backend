package url_service

import (
	"context"
	"errors"
	"log"
	"net/url"
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
	err := val.Validate.Struct(req)
	if err != nil {
		return "", shortlink_errors.ErrValidateRequest
	}

	// if CustomID is not provided, generate a new ID
	if req.CustomID == "" {
		req.CustomID, err = nanoid.New()
		if err != nil {
			log.Println("Error generating ID:", err)
			return "", shortlink_errors.ErrGenerateID
		}
	} else if err := s.validateCustomID(ctx, req.CustomID); err != nil {
		return "", err
	}

	if err := s.validateURL(ctx, req.URL); err != nil {
		return "", err
	}

	return s.saveShortlink(ctx, req)
}

func (s *URLServiceImpl) validateCustomID(ctx context.Context, customID string) error {
	if utils.BlacklistedCustomIDs[customID] { // check reserved keywords
		return shortlink_errors.ErrBlacklistedID
	}

	// check if custom ID is already exist
	_, err := s.shortlink.GetShortlink(ctx, customID)
	if err == nil {
		return shortlink_errors.ErrIDExists
	}
	if !errors.Is(err, shortlink_errors.ErrNotFound) {
		return shortlink_errors.ErrFailedRetrieveData
	}

	return nil
}

func (s *URLServiceImpl) validateURL(ctx context.Context, targetURL string) error {
	// check if the URL is valid
	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Host == "" {
		log.Println("Invalid request: URL has no host")
		return shortlink_errors.ErrValidateRequest
	}

	eg, ctx := errgroup.WithContext(ctx)

	// check if urls/its domain is blacklisted
	eg.Go(func() error {
		isBlacklisted, err := s.blacklist.IsBlacklisted(ctx, targetURL)
		if err != nil {
			log.Println("Error while checking blacklisted items:")
			return err
		}
		if isBlacklisted {
			log.Println("This URL/domain is blacklisted")
			return shortlink_errors.ErrForbiddenInput
		}
		return nil
	})

	// check safe Browsing
	eg.Go(func() error {
		isUnsafe, err := s.safebrowsing.IsUnsafe(ctx, targetURL)
		if err != nil {
			log.Printf("SafeBrowsing error: %v", err)
			return shortlink_errors.ErrFailedRetrieveData
		}
		if isUnsafe {
			log.Println("This site is unsafe")
			return shortlink_errors.ErrForbiddenInput
		}
		return nil
	})

	return eg.Wait()
}

// Simpan shortlink (reusable function)
func (s *URLServiceImpl) saveShortlink(ctx context.Context, req dto.ShortenRequest) (string, error) {
	user, ok := ctx.Value(utils.UserKey).(string)
	if !ok {
		return "", shortlink_errors.ErrValidateRequest
	}

	doc := models.Shortlink{
		ShortID:   req.CustomID,
		URL:       req.URL,
		CreatedAt: time.Now(),
		CreatedBy: user,
		IsPrivate: req.IsPrivate,
	}

	if err := s.shortlink.SetShortlink(ctx, doc.ShortID, doc); err != nil {
		return "", shortlink_errors.ErrSaveShortlink
	}
	return doc.ShortID, nil
}

func (s *URLServiceImpl) GetUserLinks(ctx context.Context, req dto.UserLinksRequest) (*dto.UserLinksResponse, error) {
	if err := val.Validate.Struct(req); err != nil {
		return nil, shortlink_errors.ErrValidateRequest
	}

	links, nextCursor, err := s.shortlink.ListUserLinks(ctx, req)
	if err != nil {
		return nil, err
	}

	dtoLinks := make([]dto.ShortlinkDTO, 0, len(links))
	for _, l := range links {
		dtoLinks = append(dtoLinks, dto.ShortlinkDTO{
			ShortID:   l.ShortID,
			URL:       l.URL,
			CreatedAt: l.CreatedAt,
			IsPrivate: l.IsPrivate,
		})
	}

	return &dto.UserLinksResponse{
		Links:      dtoLinks,
		NextCursor: nextCursor,
	}, nil
}
