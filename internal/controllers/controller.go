package controllers

import (
	"github.com/julienschmidt/httprouter"
	mw "github.com/mfmahendr/url-shortener-backend/internal/middleware"
	firestore "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	tracking "github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
)

type URLController struct {
	shortenService   url_service.URLService
	trackingService  tracking.TrackingService
	blacklistManager firestore.BlacklistManager

	Router *httprouter.Router

	RateLimiter *mw.SlidingWindowLimiter
}

func New(s url_service.URLService, t tracking.TrackingService, b firestore.BlacklistManager, l *mw.SlidingWindowLimiter) *URLController {
	return &URLController{
		shortenService:  s,
		trackingService: t,
		blacklistManager: b,
		Router:          httprouter.New(),
		RateLimiter:     l,
	}
}
