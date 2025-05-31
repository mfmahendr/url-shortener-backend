package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/shortlink_errors"
)

func mapErrorToStatusCode(err error) (statusCode int) {
	switch {
	case errors.Is(err, shortlink_errors.ErrBlacklistedID):
		statusCode = http.StatusForbidden
	case errors.Is(err, shortlink_errors.ErrIDExists):
		statusCode = http.StatusConflict
	case errors.Is(err, shortlink_errors.ErrGenerateID), errors.Is(err, shortlink_errors.ErrSaveShortlink), errors.Is(err, shortlink_errors.ErrFailedRetrieveData):
		statusCode = http.StatusInternalServerError
	case errors.Is(err, shortlink_errors.ErrValidateRequest):
		statusCode = http.StatusUnprocessableEntity
	case errors.Is(err, shortlink_errors.ErrNotFound):
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
		log.Printf("Unexpected error: %v", err)
	}
	return statusCode
}
