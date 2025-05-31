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

func New(s url_service.URLService) *URLController {
	return &URLController{
		ShortenService: s,
		Router:         httprouter.New(),
	}
}
