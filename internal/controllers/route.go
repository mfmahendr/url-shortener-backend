package controllers

import (
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
)

func (c *URLController) RegisterRoutes(authMiddleware middleware.AuthMiddleware) {
	c.Router.GET("/", c.Home)
	c.Router.GET("/health", c.HealthCheck)

	c.Router.POST("/shorten", authMiddleware.RequireAuth(c.Shorten))
	c.Router.GET("/:short_id", c.Redirect)

	c.Router.GET("/:short_id/click-count", authMiddleware.RequireAuth(c.GetClickCount))
	c.Router.GET("/:short_id/analytics", authMiddleware.RequireAuth(c.Analytics))

}
