package safebrowsing_service

import (
	"context"
	"log"
)

type MockSafeBrowsingService struct {
	UnsafeURLs map[string]bool
	Err        error
}

func (m *MockSafeBrowsingService) IsUnsafe(ctx context.Context, targetURL string) (bool, error) {
	if m.Err != nil {
		return false, m.Err
	}
	log.Println("[MockSafeBrowsingService] IsUnsafe ("+targetURL+")?", m.UnsafeURLs[targetURL])
	return m.UnsafeURLs[targetURL], nil
}
