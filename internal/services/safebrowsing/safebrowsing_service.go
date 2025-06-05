package safebrowsing_service

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"google.golang.org/api/option"
	"google.golang.org/api/safebrowsing/v4"
)

type URLSafetyChecker interface {
	IsUnsafe(ctx context.Context, targetURL string) (bool, error)
}

type SafeBrowsingServiceImpl struct {
	apiKey  string
	Service *safebrowsing.Service
}

func New(ctx context.Context, apiKey string) URLSafetyChecker {
	service, _ := safebrowsing.NewService(ctx, option.WithAPIKey(apiKey))

	return &SafeBrowsingServiceImpl{
		apiKey:  apiKey,
		Service: service,
	}
}

func (s *SafeBrowsingServiceImpl) IsUnsafe(ctx context.Context, targetURL string) (bool, error) {
	if err := validators.Validate.Var(targetURL, "required,url"); err != nil {
		return false, shortlink_errors.ErrValidateRequest
	}

	req := &safebrowsing.GoogleSecuritySafebrowsingV4FindThreatMatchesRequest{
		Client: &safebrowsing.GoogleSecuritySafebrowsingV4ClientInfo{
			ClientId:      "url-shortener",
			ClientVersion: "1.0",
		},
		ThreatInfo: &safebrowsing.GoogleSecuritySafebrowsingV4ThreatInfo{
			ThreatTypes:      []string{"MALWARE", "SOCIAL_ENGINEERING"},
			PlatformTypes:    []string{"ANY_PLATFORM"},
			ThreatEntryTypes: []string{"URL"},
			ThreatEntries: []*safebrowsing.GoogleSecuritySafebrowsingV4ThreatEntry{
				{Url: targetURL},
			},
		},
	}

	response, err := s.Service.ThreatMatches.Find(req).Do()
	if err != nil {
		return false, err
	}

	if len(response.Matches) > 0 {
		return true, nil
	}

	return false, nil
}
