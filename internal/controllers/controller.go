package controllers

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
)

type URLController struct {
	Service url_service.URLService
	Router  *httprouter.Router
}

func New(s url_service.URLService) *URLController {
	return &URLController{
		Service: s,
		Router:  httprouter.New(),
	}
}
