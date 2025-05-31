package controllers

import (
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
)

func (c *URLController) RegisterRoutes(authMiddleware middleware.AuthMiddleware) {
	c.Router.POST("/shorten", authMiddleware.RequireAuth(c.Shorten))
	c.Router.GET("/:short_id", c.Redirect)

	c.Router.GET("/click-count/:short_id", authMiddleware.RequireAuth(c.GetClickCount))
	c.Router.GET("/analytics/:short_id", authMiddleware.RequireAuth(c.Analytics))

}
