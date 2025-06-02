package dto

import (
	"time"
)

type AnalyticsDTO struct {
	ShortID     string        `json:"short_id"`
	TotalClicks int64         `json:"total_clicks"`
	Clicks      []ClickLogDTO `json:"clicks"`
	NextCursor  string        `json:"next_cursor"`
}

type ClickLogDTO struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
}
