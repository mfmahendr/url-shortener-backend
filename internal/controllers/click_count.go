package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)

func (c *URLController) GetClickCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	shortID := ps.ByName("short_id")

	count, err := c.TrackingService.GetClickCount(ctx, shortID)
	if err != nil {
		http.Error(w, "Failed to get click count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	clickCountResponse := dto.ClickCountResponse{
		ShortID:    shortID,
		ClickCount: count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&clickCountResponse)
}
