package controllers

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (c *URLController) AnalyticsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	shortID := ps.ByName("short_id")

	responseData, err := c.TrackingService.GetAnalytics(ctx, shortID)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to fetch analytics", statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
