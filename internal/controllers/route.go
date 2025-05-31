package controllers

import (
	"net/http"

	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
)


func (c *URLController) RegisterRoutes(authMiddleware middleware.AuthMiddleware) {
	c.Router.Handle(http.MethodPost, "/shorten", authMiddleware.RequireAuth(c.Shorten))
	c.Router.Handle(http.MethodGet, "/:short_id", c.Redirect)

	c.Router.Handle(http.MethodGet, "/:short_id/click-count", c.GetClickCount)
}
