package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
)

func (c *URLController) Analytics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// authenticate user access and ownership
	ctx := r.Context()
	shortID := ps.ByName("short_id")

	user, ok := ctx.Value(utils.UserKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isOwner, err := c.shortenService.IsOwner(ctx, shortID, user)
	if verifyOwnerAccess(w, err, isOwner) {
		return
	}

	// Parse req params
	req := &dto.ClickLogsRequest{
		ShortID: shortID,
	}
	parseClickLogsQuery(r, &req.ClickLogsQuery)

	// Fetch analytics data
	responseData, err := c.trackingService.GetAnalytics(ctx, *req)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to fetch analytics", statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
