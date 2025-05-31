package controllers

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (c *URLController) Redirect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	shortID := p.ByName("short_id")

	
	url, err := c.Service.Resolve(ctx, shortID)
	if err != nil {
		log.Printf("Error resolving short ID %s: %v", shortID, err)
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}