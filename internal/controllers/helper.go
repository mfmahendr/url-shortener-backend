package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)

func mapErrorToStatusCode(err error) (statusCode int) {
	switch {
	case errors.Is(err, shortlink_errors.ErrBlacklistedID), errors.Is(err, shortlink_errors.ErrForbidden):
		statusCode = http.StatusForbidden
	case errors.Is(err, shortlink_errors.ErrIDExists):
		statusCode = http.StatusConflict
	case errors.Is(err, shortlink_errors.ErrGenerateID), errors.Is(err, shortlink_errors.ErrSaveShortlink), errors.Is(err, shortlink_errors.ErrFailedRetrieveData):
		statusCode = http.StatusInternalServerError
	case errors.Is(err, shortlink_errors.ErrValidateRequest):
		statusCode = http.StatusBadRequest
	case errors.Is(err, shortlink_errors.ErrNotFound):
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
		log.Printf("Unexpected error: %v", err)
	}
	return statusCode
}

func verifyOwnerAccess(w http.ResponseWriter, err error, isOwner bool) bool {
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to check ownership: "+err.Error(), statusCode)
		return true
	}
	if !isOwner {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return true
	}
	return false
}

func parseClickLogsQuery(r *http.Request, query *dto.ClickLogsQuery) {
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			query.Limit = parsed
		}
	} else {
		query.Limit = 50
	}

	query.Cursor = r.URL.Query().Get("cursor")
	query.OrderDesc = r.URL.Query().Get("order") == "desc"

	if after := r.URL.Query().Get("after"); after != "" {
		t, _ := time.Parse(time.RFC3339, after)
		query.After = t
	}
	if before := r.URL.Query().Get("before"); before != "" {
		t, _ := time.Parse(time.RFC3339, before)
		query.Before = t
	}
}
