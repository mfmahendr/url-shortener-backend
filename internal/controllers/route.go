package controllers

import (
	mw "github.com/mfmahendr/url-shortener-backend/internal/middleware"
)

func (c *URLController) RegisterRoutes(auth mw.AuthMiddleware) {
	c.Router.GET("/health", c.RateLimiter.Apply(c.HealthCheck))
	c.Router.GET("/", c.RateLimiter.Apply(c.Home))
	
	c.Router.GET("/r/:short_id", c.RateLimiter.Apply(c.Redirect))
	
	c.Router.POST("/u/shorten", c.RateLimiter.Apply(auth.RequireAuth(c.Shorten)))
	c.Router.GET("/u/click-count/:short_id", c.RateLimiter.Apply(auth.RequireAuth(c.GetClickCount)))
	c.Router.GET("/u/analytics/:short_id", c.RateLimiter.Apply(auth.RequireAuth(c.Analytics)))

	// admin
	c.Router.GET("/admin/blacklist", c.RateLimiter.Apply(auth.RequireAdminAuth(c.FetchBlacklistedDomains)))
	c.Router.POST("/admin/blacklist", c.RateLimiter.Apply(auth.RequireAdminAuth(c.AddToBlacklist)))
	c.Router.DELETE("/admin/blacklist", c.RateLimiter.Apply(auth.RequireAdminAuth(c.RemoveFromBlacklist)))
}
