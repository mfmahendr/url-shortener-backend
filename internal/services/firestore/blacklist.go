package firestore_service

import (
	"context"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	BlacklistManager interface {
		BlacklistDomain(ctx context.Context, domain string) error
		UnblacklistDomain(ctx context.Context, domain string) error
		ListBlacklistedDomains(ctx context.Context) ([]string, error)
	}

	BlacklistChecker interface {
		IsDomainBlacklisted(ctx context.Context, domain string) (bool, error)
	}
)

// Blacklist manager implementation
func (s *FirestoreServiceImpl) BlacklistDomain(ctx context.Context, domain string) error {
	if err := validators.Validate.Var(domain, "required,hostname"); err != nil {
		return shortlink_errors.ErrValidateRequest
	}

	ok, err := s.isDomainBlacklisted(ctx, domain)
	if err != nil {
		return err
	}
	if ok {
		return shortlink_errors.ErrResourceExists
	}

	_, err = s.client.Collection("blacklist_domains").Doc(domain).Set(ctx, map[string]interface{}{
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

	_, err = s.client.Collection("blacklist_domains").Doc(domain).Delete(ctx)
	return err
}

func (s *FirestoreServiceImpl) ListBlacklistedDomains(ctx context.Context) ([]string, error) {
	iter := s.client.Collection("blacklist_domains").Documents(ctx)
	var list []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, shortlink_errors.ErrFailedRetrieveData
		}
		list = append(list, doc.Ref.ID)
	}
	return list, nil
}

func (s *FirestoreServiceImpl) IsDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	if err := validators.Validate.Var(domain, "required,hostname"); err != nil {
		return false, shortlink_errors.ErrValidateRequest
	}

	return s.isDomainBlacklisted(ctx, domain)
}

func (s *FirestoreServiceImpl) isDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	doc, err := s.client.Collection("blacklist_domains").Doc(domain).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, shortlink_errors.ErrFailedRetrieveData
	}
	return doc.Exists(), nil
}
