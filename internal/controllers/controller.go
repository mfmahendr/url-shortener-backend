package controllers

import (
	"github.com/julienschmidt/httprouter"
	mw "github.com/mfmahendr/url-shortener-backend/internal/middleware"
	tracking "github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
)

type URLController struct {
	shortenService  url_service.URLService
	trackingService tracking.TrackingService
	Router          *httprouter.Router

	RateLimiter *mw.SlidingWindowLimiter
}

func New(shorten url_service.URLService, tracking tracking.TrackingService, limiter *mw.SlidingWindowLimiter) *URLController {
	return &URLController{
		shortenService:  shorten,
		trackingService: tracking,
		Router:          httprouter.New(),
		RateLimiter:     limiter,
	}
}
