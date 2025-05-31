package controllers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"
)

type Controller interface {
	Shorten(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	Redirect(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	RouteRegistrar
}

type URLController struct {
	Service url_service.URLService
	Router  *httprouter.Router
}

func New(s url_service.URLService) Controller {
	return &URLController{
		Service: s,
		Router:  httprouter.New(),
	}
}
