package safebrowsing_service

import (
	"context"
	"log"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
	"google.golang.org/api/safebrowsing/v4"
)

type URLSafetyChecker interface {
	IsUnsafe(ctx context.Context, targetURL string) (bool, error)
}

type SafeBrowsingServiceImpl struct {
	apiKey  string
	service *safebrowsing.Service
	redis   *redis.Client
}

func New(ctx context.Context, apiKey string, redis *redis.Client) URLSafetyChecker {
	service, _ := safebrowsing.NewService(ctx, option.WithAPIKey(apiKey))

	return &SafeBrowsingServiceImpl{
		apiKey:  apiKey,
		service: service,
		redis: redis,
	}
}

func (s *SafeBrowsingServiceImpl) IsUnsafe(ctx context.Context, targetURL string) (bool, error) {
	if err := validators.Validate.Var(targetURL, "required,url"); err != nil {
		return false, shortlink_errors.ErrValidateRequest
	}

	// check cache
	cacheKey := "safebrowsing:" + targetURL
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Println("Check safebrowsing")
		return cached == "unsafe", nil
	}

	// check API
	unsafe, err := s.requestSafeBrowsingChecking(targetURL)
	if err != nil {
		return false, err
	}
	
	// cache the URL for 24 hours
	cacheVal := "safe"
	if unsafe {
		cacheVal = "unsafe"
	}
	_ = s.redis.Set(ctx, cacheKey, cacheVal, 24*time.Hour).Err()	// expires in 24 hours

	return unsafe, nil
}

func (s *SafeBrowsingServiceImpl) requestSafeBrowsingChecking(targetURL string) (bool, error) {
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

	response, err := s.service.ThreatMatches.Find(req).Do()
	if err != nil {
		return false, err
	}

	if len(response.Matches) > 0 {
		return true, nil
	}

	return false, nil
}
