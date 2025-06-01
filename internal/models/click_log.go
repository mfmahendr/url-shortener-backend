package models

import "time"

type ClickLog struct {
	ShortID    string    `json:"short_id,omitempty" firestore:"short_id"`
	Timestamp  time.Time `json:"timestamp" firestore:"timestamp"`
	IP         string    `json:"ip" firestore:"ip"`
	UserAgent  string    `json:"user_agent" firestore:"user_agent"`
}
