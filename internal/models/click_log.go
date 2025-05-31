package models

import "time"

type ClickLog struct {
	Timestamp  time.Time `json:"timestamp"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
}
