package controllers

import (
	"net/http"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)

func (c *URLController) Shorten(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	_, ok := ctx.Value(utils.UserKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Failed to shorten URL:" + shortlink_errors.ErrValidateRequest.Error(), http.StatusBadRequest)
		return
	}

	shortID, err := c.shortenService.Shorten(ctx, req)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to shorten URL: "+err.Error(), statusCode)
		return
	}

	response := dto.ShortenResponse{ShortID: shortID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
