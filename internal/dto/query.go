package dto

import (
	"time"
)

type PaginationQuery struct {
	Limit     int    `json:"limit" validate:"omitempty,min=1"`
	Cursor    string `json:"cursor" validate:"omitempty"`
	OrderDesc bool   `json:"order_desc" validate:"-"`
}

type ClickLogsQuery struct {
	UserAgent string    `json:"user_agent,omitempty" validate:"omitempty"`
	After     time.Time `json:"after" validate:"omitempty,datetime"`
	Before    time.Time `json:"before" validate:"omitempty,datetime"`
	PaginationQuery
}
