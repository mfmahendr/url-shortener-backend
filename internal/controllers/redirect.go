package controllers

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
	"log"

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

    ip := "unknown"
    if remote := r.RemoteAddr; remote != "" {
        if host, _, err := net.SplitHostPort(remote); err == nil {
            ip = host
        } else {
            ip = remote
        }
    }
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        ip = strings.Split(forwarded, ",")[0]
    }
    ua := r.UserAgent()

    go func(shortID, ip, ua string) {
        trackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := c.trackingService.TrackClick(trackCtx, shortID, ip, ua); err != nil {
            log.Printf("TrackClick failed for %s: %v", shortID, err)
        } else {
            log.Printf("TrackClick success for %s", shortID)
        }
    }(shortID, ip, ua)

    http.Redirect(w, r, url, http.StatusFound)
}
