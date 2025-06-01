package controllers

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
)

type URLController struct {
	ShortenService  url_service.URLService
	TrackingService tracking_service.TrackingService
	Router          *httprouter.Router
}

func New(shorten url_service.URLService, tracking tracking_service.TrackingService) *URLController {
	return &URLController{
		ShortenService:  shorten,
		TrackingService: tracking,
		Router:          httprouter.New(),
	}
}
