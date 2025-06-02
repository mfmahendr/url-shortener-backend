package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)

func (c *URLController) Analytics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// authenticate user access and ownership
	ctx := r.Context()
	shortID := ps.ByName("short_id")

	user, ok := ctx.Value("user").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isOwner, err := c.ShortenService.IsOwner(ctx, shortID, user)
	if verifyOwnerAccess(w, err, isOwner) {
		return
	}

	// Parse query params
	query := &dto.ClickLogsQuery{
		ShortID: shortID,
	}
	parseClickLogsQuery(r, query)

	// Fetch analytics data
	responseData, err := c.TrackingService.GetAnalytics(ctx, *query)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to fetch analytics", statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
