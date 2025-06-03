package controllers

import (
	mw "github.com/mfmahendr/url-shortener-backend/internal/middleware"
)

func (c *URLController) RegisterRoutes(auth mw.AuthMiddleware) {
	c.Router.GET("/", c.RateLimiter.Apply(c.Home))
	c.Router.GET("/health", c.RateLimiter.Apply(c.HealthCheck))

	c.Router.POST("/shorten", c.RateLimiter.Apply(auth.RequireAuth(c.Shorten)))
	c.Router.GET("/:short_id", c.RateLimiter.Apply(c.Redirect))

	c.Router.GET("/:short_id/click-count", c.RateLimiter.Apply(auth.RequireAuth(c.GetClickCount)))
	c.Router.GET("/:short_id/analytics", c.RateLimiter.Apply(auth.RequireAuth(c.Analytics)))

}
