package url_service

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	firestore "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	safebrowsing "github.com/mfmahendr/url-shortener-backend/internal/services/safebrowsing"
)

type URLService interface {
	Shorten(ctx context.Context, req dto.ShortenRequest) (shortID string, err error)
	Resolve(ctx context.Context, shortID string) (string, error)
	IsOwner(ctx context.Context, shortID string, uid string) (bool, error)
	GetUserLinks(ctx context.Context, req dto.UserLinksRequest) (*dto.UserLinksResponse, error)
}

type URLServiceImpl struct {
	shortlink    firestore.Shortlink
	blacklist    firestore.BlacklistChecker
	safebrowsing safebrowsing.URLSafetyChecker
	// safebrowsing *safebrowsing.Service
}

func New(sl firestore.Shortlink, bl firestore.BlacklistChecker, sb safebrowsing.URLSafetyChecker) URLService {
	return &URLServiceImpl{
		shortlink:    sl,
		blacklist:    bl,
		safebrowsing: sb,
	}
}
