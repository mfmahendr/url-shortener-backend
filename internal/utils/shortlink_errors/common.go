package shortlink_errors

import "errors"

var (
	ErrValidateRequest    = errors.New("invalid request data")
	ErrNotFound           = errors.New("no data found")
	ErrForbidden          = errors.New("forbidden access")
)
