package controllers

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (c *URLController) Redirect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()
	shortID := p.ByName("short_id")

	url, err := c.shortenService.Resolve(ctx, shortID)
	if err != nil {
		log.Printf("Error resolving short ID %s: %v", shortID, err)
		http.Error(w, err.Error(), mapErrorToStatusCode(err))
		return
	}

	// Tracking click
	go func(ctx context.Context, r *http.Request, shortID string) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Unable to parse RemoteAddr: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}
		ua := r.UserAgent()
		c.trackingService.TrackClick(ctx, shortID, ip, ua)
	}(ctx, r, shortID)

	http.Redirect(w, r, url, http.StatusFound)
}
