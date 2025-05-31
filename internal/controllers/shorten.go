package controllers

import (
	"context"
	"net/http"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)

func (c *URLController) Shorten(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := context.Background()

	var req dto.ShortenerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortID, err := c.Service.Shorten(ctx, req)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to shorten URL: "+err.Error(), statusCode)
		return
	}

	response := dto.ShortenerResponse{ShortID: shortID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
