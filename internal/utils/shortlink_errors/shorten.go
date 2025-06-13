package shortlink_errors

import "errors"

var (
	ErrSaveShortlink      = errors.New("failed to save short link")
	ErrFailedRetrieveData = errors.New("failed to retrieve data from database")
)
