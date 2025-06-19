package safebrowsing_service

import (
	"context"
)

type MockSafeBrowsingService struct {
	UnsafeURLs map[string]bool
	Err        error
}

func (m *MockSafeBrowsingService) IsUnsafe(ctx context.Context, targetURL string) (bool, error) {
	if m.Err != nil {
		return false, m.Err
	}
	return m.UnsafeURLs[targetURL], nil
}
