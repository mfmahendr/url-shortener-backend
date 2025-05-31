package controllers

import (
	"net/http"
)


type RouteRegistrar interface {
	RegisterRoutes()
}

func (c *URLController) RegisterRoutes() {
	c.Router.Handle(http.MethodPost, "/shorten", c.Shorten)
	c.Router.Handle(http.MethodGet, "/:short_id", c.Redirect)
}
