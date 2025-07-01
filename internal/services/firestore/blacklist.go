package firestore_service

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/models"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	BlacklistChecker interface {
		IsBlacklisted(ctx context.Context, url string) (bool, error)
	}

	BlacklistManager interface {
		BlacklistDomain(ctx context.Context, domain string) error
		BlacklistURL(ctx context.Context, inputURL string) error
		UnblacklistDomain(ctx context.Context, domain string) error
		UnblacklistURL(ctx context.Context, inputURL string) error
		ListBlacklisted(ctx context.Context) ([]models.BlacklistItem, error)
}
)

// Blacklist manager implementation
func (s *FirestoreServiceImpl) BlacklistDomain(ctx context.Context, domain string) error {
	if err := validators.Validate.Var(domain, "required,hostname"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}

	exists, err := s.isDomainBlacklisted(ctx, domain)
	if err != nil {
		return err
	}
	if exists {
		return shortlink_errors.ErrResourceExists
	}

	_, err = s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(domain)).Set(ctx, map[string]interface{}{
		"type":       "domain",
		"value":      domain,
		"created_at": time.Now(),
		"source":     "manual",
	})

	return err
}

func (s *FirestoreServiceImpl) BlacklistURL(ctx context.Context, inputURL string) error {
	parsed, err := url.Parse(inputURL)
	if err != nil || parsed.Host == "" {
		return shortlink_errors.ErrValidateRequest
	}

	normalizedURL := normalizeURLForBlacklist(parsed)

	exists, err := s.isURLBlacklisted(ctx, inputURL)
	if err != nil {
		return err
	}
	if exists {
		return shortlink_errors.ErrResourceExists
	}

	_, err = s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(normalizedURL)).Set(ctx, map[string]interface{}{
		"type":       "url",
		"value":      normalizedURL,
		"created_at": time.Now(),
		"source":     "manual",
	})
	return err
}


func (s *FirestoreServiceImpl) UnblacklistDomain(ctx context.Context, domain string) error {
	if err := validators.Validate.Var(domain, "required,hostname"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}

	ok, err := s.isDomainBlacklisted(ctx, domain)
	if err != nil {
		return err
	}
	if !ok {
		return shortlink_errors.ErrNotFound
	}

	_, err = s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(domain)).Delete(ctx)
	return err
}

func (s *FirestoreServiceImpl) UnblacklistURL(ctx context.Context, inputURL string) error {
	if err := validators.Validate.Var(inputURL, "required,url"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}
	
	parsed, err := url.Parse(inputURL)
	if err != nil || parsed.Host == "" {
		return shortlink_errors.ErrValidateRequest
	}

	normalizedURL := normalizeURLForBlacklist(parsed)

	doc, err := s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(normalizedURL)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return shortlink_errors.ErrNotFound
		}
		return shortlink_errors.ErrFailedRetrieveData
	}

	if !doc.Exists() {
		return shortlink_errors.ErrNotFound
	}

	_, err = s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(normalizedURL)).Delete(ctx)
	return err
}

func (s *FirestoreServiceImpl) ListBlacklisted(ctx context.Context) ([]models.BlacklistItem, error) {
	iter := s.client.Collection("blacklist_items").Documents(ctx)
	var list []models.BlacklistItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("Unexpected error:", err)
			return nil, shortlink_errors.ErrFailedRetrieveData
		}
		
		var item models.BlacklistItem
		if err := doc.DataTo(&item); err != nil {
			log.Println("Unexpected error:", err)
			return nil, shortlink_errors.ErrFailedRetrieveData
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *FirestoreServiceImpl) IsBlacklisted(ctx context.Context, inputURL string) (bool, error) {
	if inputURL == "" {
		return false, shortlink_errors.ErrValidateRequest
	}

	parsed, err := url.Parse(inputURL)
	if err != nil || parsed.Host == "" {
		if err := validators.Validate.Var(inputURL, "hostname"); err != nil {
			return false, shortlink_errors.ErrValidateRequest
		}
		return s.isDomainBlacklisted(ctx, inputURL)
	}

	domain := parsed.Hostname()

	isDomainBlacklisted, err := s.isDomainBlacklisted(ctx, domain)
	if err != nil {
		return false, err
	}

	isURLBlacklisted, err := s.isURLBlacklisted(ctx, inputURL)
	if err != nil {
		return false, err
	}

	return isDomainBlacklisted || isURLBlacklisted, nil
}

func (s *FirestoreServiceImpl) isDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	doc, err := s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(domain)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, shortlink_errors.ErrFailedRetrieveData
	}
	return doc.Exists(), nil
}

func (s *FirestoreServiceImpl) isURLBlacklisted(ctx context.Context, inputURL string) (bool, error) {
	parsed, _ := url.Parse(inputURL)
	normalizedURL := normalizeURLForBlacklist(parsed)
	
	doc, err := s.client.Collection("blacklist_items").Doc(utils.GenerateDocID(normalizedURL)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, shortlink_errors.ErrFailedRetrieveData
	}
	return doc.Exists(), nil
}
