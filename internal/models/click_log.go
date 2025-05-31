package models

import "time"

type ClickLog struct {
	ShortID    string    `json:"short_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
}
