package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
)

func (c *URLController) GetClickCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	count, err := c.trackingService.GetClickCount(ctx, shortID)
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

func (c *URLController) ExportAllClickCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shortID := ps.ByName("short_id")

	// check ownership
	user, ok := r.Context().Value(utils.UserKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isOwner, err := c.shortenService.IsOwner(r.Context(), shortID, user)
	if verifyOwnerAccess(w, err, isOwner) {
		return
	}

	// get format
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}
	ctx := context.WithValue(r.Context(), utils.ExportFormatKey, format)
	query := dto.ClickLogsQuery{
		ShortID: shortID,
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analytics_%s.csv\"", shortID))
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analytics_%s.json\"", shortID))
	default:
		http.Error(w, "Unsupported format; use ?format=csv or ?format=json", http.StatusBadRequest)
		return
	}

	// Stream click logs
	err = c.trackingService.StreamClickLogs(ctx, w, query)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, err.Error(), statusCode)
		return
	}
}
