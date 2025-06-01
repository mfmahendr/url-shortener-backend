package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (c *URLController) Redirect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	shortID := p.ByName("short_id")

	url, err := c.ShortenService.Resolve(ctx, shortID)
	if err != nil {
		log.Printf("Error resolving short ID %s: %v", shortID, err)
		http.NotFound(w, r)
		return
	}

	// Tracking click
	go func(ctx context.Context, r *http.Request, shortID string) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}
		ua := r.UserAgent()
		c.TrackingService.TrackClick(ctx, shortID, ip, ua)
	}(ctx, r, shortID)

	http.Redirect(w, r, url, http.StatusFound)
}
